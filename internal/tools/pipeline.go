// Package tools provides the pipeline executor for chaining capture operations.
//
// The pipeline executor implements a stack-based execution model where each
// tool step pushes its result onto a stack. Subsequent steps can consume
// items from the stack as input.
//
// Stack behavior per tool:
//
//   - capture_screen: pushes image (raw bytes)
//   - capture_window: pushes image (raw bytes)
//   - capture_region: pushes image (raw bytes); pops region if no explicit coords
//   - find_region: pops 1 image, pushes text (JSON coords)
//   - extract_text: pops 1 image, pushes text
//   - analyze_image: pops 1 image, pushes text
//   - compare_images: pops 2 images, pushes text
//   - wait_for: no input, no output
//
// At the end of the pipeline, only the top stack item is returned.
// Unused items are discarded.
package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/logging"
)

// PipelineStep represents a single step in a capture pipeline.
//
// Each step specifies which tool to execute and provides the parameters
// for that tool. The tool's output is pushed onto the pipeline stack
// for use by subsequent steps.
type PipelineStep struct {
	// Tool is the name of the tool to execute.
	// Valid values: capture_screen, capture_window, capture_region,
	// find_region, extract_text, analyze_image, compare_images, wait_for.
	Tool string `json:"tool"`

	// Parameters contains the tool-specific parameters.
	// The exact fields depend on the tool being executed.
	Parameters map[string]any `json:"parameters,omitempty"`
}

// pipelineStack implements a simple stack for pipeline execution.
type pipelineStack struct {
	items []any
}

// push adds an item to the top of the stack.
func (s *pipelineStack) push(item any) {
	s.items = append(s.items, item)
}

// pop removes and returns the top item from the stack.
// Returns an error if the stack is empty.
func (s *pipelineStack) pop() (any, error) {
	if len(s.items) == 0 {
		return nil, fmt.Errorf("stack is empty")
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, nil
}

// popImage pops an item from the stack and returns it as raw image bytes.
func (s *pipelineStack) popImage() ([]byte, error) {
	item, err := s.pop()
	if err != nil {
		return nil, fmt.Errorf("expected image on stack: %w", err)
	}
	data, ok := item.([]byte)
	if !ok {
		return nil, fmt.Errorf("expected image on stack, got %T", item)
	}
	return data, nil
}

// popRegion pops a region from the stack and returns x, y, width, height.
// The region can be either a map with x/y/width/height keys or a JSON string.
func (s *pipelineStack) popRegion() (x, y, w, h int, err error) {
	item, err := s.pop()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("expected region on stack: %w", err)
	}

	switch v := item.(type) {
	case map[string]any:
		x = int(v["x"].(float64))
		y = int(v["y"].(float64))
		w = int(v["width"].(float64))
		h = int(v["height"].(float64))
	case string:
		var region struct {
			X      int `json:"x"`
			Y      int `json:"y"`
			Width  int `json:"width"`
			Height int `json:"height"`
		}
		if err := json.Unmarshal([]byte(v), &region); err != nil {
			return 0, 0, 0, 0, fmt.Errorf("failed to parse region from stack: %w", err)
		}
		x = region.X
		y = region.Y
		w = region.Width
		h = region.Height
	default:
		return 0, 0, 0, 0, fmt.Errorf("expected region on stack, got %T", item)
	}

	return x, y, w, h, nil
}

// ExecutePipeline executes a pipeline of capture steps.
//
// Each step is executed in order. The step's output is pushed onto a stack
// for use by subsequent steps. At the end of the pipeline, the top stack
// item is returned. Unused items are discarded.
//
// The tools argument provides the capture and vision tool implementations
// that the pipeline steps delegate to.
//
// Returns the result as either a base64-encoded image string (imageBase64)
// or a text string (text). Only one of the two will be non-empty.
func ExecutePipeline(ctx context.Context, steps []PipelineStep, t *Tools) (imageBase64 string, text string, err error) {
	stack := &pipelineStack{}

	for i, step := range steps {
		logging.Debug().Int("step", i+1).Str("tool", step.Tool).Msg("Executing pipeline step")

		var err error
		switch step.Tool {
		case "capture_screen":
			err = execCaptureScreen(ctx, step.Parameters, t, stack)
		case "capture_window":
			err = execCaptureWindow(ctx, step.Parameters, t, stack)
		case "capture_region":
			err = execCaptureRegion(ctx, step.Parameters, t, stack)
		case "find_region":
			err = execFindRegion(ctx, step.Parameters, t, stack)
		case "extract_text":
			err = execExtractText(ctx, step.Parameters, t, stack)
		case "analyze_image":
			err = execAnalyzeImage(ctx, step.Parameters, t, stack)
		case "compare_images":
			err = execCompareImages(ctx, step.Parameters, t, stack)
		case "wait_for":
			err = execWaitFor(step.Parameters)
		default:
			return "", "", fmt.Errorf("unknown pipeline tool: %q", step.Tool)
		}

		if err != nil {
			return "", "", fmt.Errorf("pipeline step %d (%s) failed: %w", i+1, step.Tool, err)
		}
	}

	if len(stack.items) == 0 {
		return "", "", fmt.Errorf("pipeline produced no output")
	}

	// Return only the top item, discard the rest
	result := stack.items[len(stack.items)-1]

	// If it's image data, return as base64
	if img, ok := result.([]byte); ok {
		return base64.StdEncoding.EncodeToString(img), "", nil
	}

	text, ok := result.(string)
	if !ok {
		return "", "", fmt.Errorf("unexpected result type: %T", result)
	}
	return "", text, nil
}

func execCaptureScreen(ctx context.Context, params map[string]any, t *Tools, stack *pipelineStack) error {
	monitor := ""
	if m, ok := params["monitor"].(string); ok {
		monitor = m
	}
	img, err := t.CaptureScreen(ctx, monitor)
	if err != nil {
		return err
	}
	stack.push(img)
	return nil
}

func execCaptureWindow(ctx context.Context, params map[string]any, t *Tools, stack *pipelineStack) error {
	title, ok := params["title"].(string)
	if !ok {
		return fmt.Errorf("capture_window requires 'title' parameter")
	}
	img, err := t.CaptureWindow(ctx, title)
	if err != nil {
		return err
	}
	stack.push(img)
	return nil
}

func execCaptureRegion(ctx context.Context, params map[string]any, t *Tools, stack *pipelineStack) error {
	var x, y, w, h int

	if xVal, ok := params["x"].(float64); ok {
		x = int(xVal)
		y = int(params["y"].(float64))
		w = int(params["width"].(float64))
		h = int(params["height"].(float64))
	} else {
		var err error
		x, y, w, h, err = stack.popRegion()
		if err != nil {
			return fmt.Errorf("capture_region requires explicit coordinates or a region on the stack: %w", err)
		}
	}

	img, err := t.CaptureRegion(ctx, x, y, w, h)
	if err != nil {
		return err
	}
	stack.push(img)
	return nil
}

func execFindRegion(ctx context.Context, params map[string]any, t *Tools, stack *pipelineStack) error {
	description, ok := params["description"].(string)
	if !ok {
		return fmt.Errorf("find_region requires 'description' parameter")
	}
	provider := ""
	if p, ok := params["provider"].(string); ok {
		provider = p
	}
	timeout := 0
	if tv, ok := params["timeout"].(float64); ok {
		timeout = int(tv)
	}

	img, err := stack.popImage()
	if err != nil {
		return err
	}

	result, err := t.FindRegion(ctx, img, description, provider, timeout)
	if err != nil {
		return err
	}
	stack.push(result)
	return nil
}

func execExtractText(ctx context.Context, params map[string]any, t *Tools, stack *pipelineStack) error {
	provider := ""
	if p, ok := params["provider"].(string); ok {
		provider = p
	}
	timeout := 0
	if tv, ok := params["timeout"].(float64); ok {
		timeout = int(tv)
	}

	img, err := stack.popImage()
	if err != nil {
		return err
	}

	result, err := t.ExtractText(ctx, img, provider, timeout)
	if err != nil {
		return err
	}
	stack.push(result)
	return nil
}

func execAnalyzeImage(ctx context.Context, params map[string]any, t *Tools, stack *pipelineStack) error {
	prompt, ok := params["prompt"].(string)
	if !ok {
		return fmt.Errorf("analyze_image requires 'prompt' parameter")
	}
	provider := ""
	if p, ok := params["provider"].(string); ok {
		provider = p
	}
	timeout := 0
	if tv, ok := params["timeout"].(float64); ok {
		timeout = int(tv)
	}

	img, err := stack.popImage()
	if err != nil {
		return err
	}

	result, err := t.AnalyzeImage(ctx, img, prompt, provider, timeout)
	if err != nil {
		return err
	}
	stack.push(result)
	return nil
}

func execCompareImages(ctx context.Context, params map[string]any, t *Tools, stack *pipelineStack) error {
	provider := ""
	if p, ok := params["provider"].(string); ok {
		provider = p
	}
	timeout := 0
	if tv, ok := params["timeout"].(float64); ok {
		timeout = int(tv)
	}

	img2, err := stack.popImage()
	if err != nil {
		return err
	}
	img1, err := stack.popImage()
	if err != nil {
		return err
	}

	prompt := "Describe the differences between these two images. Be specific about what changed."
	if p, ok := params["prompt"].(string); ok && p != "" {
		prompt = p
	}

	result, err := t.CompareImages(ctx, img1, img2, prompt, provider, timeout)
	if err != nil {
		return err
	}
	stack.push(result)
	return nil
}

func execWaitFor(params map[string]any) error {
	seconds, ok := params["seconds"].(float64)
	if !ok {
		return fmt.Errorf("wait_for requires 'seconds' parameter")
	}
	if seconds < 0 || seconds > 30 {
		return fmt.Errorf("wait_for seconds must be between 0 and 30, got %v", seconds)
	}
	time.Sleep(time.Duration(seconds * float64(time.Second)))
	return nil
}
