package tools

import (
	"context"
	"image"
	"testing"

	"github.com/emmanuel-deloget/screenshooter-mcp/internal/vision"
)

// mockProvider implements vision.Provider for testing
type mockProvider struct {
	name    string
	model   string
	analyze func(ctx context.Context, image []byte, prompt string) (string, error)
	compare func(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error)
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) ModelName() string {
	return m.model
}

func (m *mockProvider) Analyze(ctx context.Context, image []byte, prompt string) (string, error) {
	if m.analyze != nil {
		return m.analyze(ctx, image, prompt)
	}
	return "mock response", nil
}

func (m *mockProvider) CompareImages(ctx context.Context, image1 []byte, image2 []byte, prompt string) (string, error) {
	if m.compare != nil {
		return m.compare(ctx, image1, image2, prompt)
	}
	return "mock comparison", nil
}

func newMockVisionManager(providers ...*mockProvider) *vision.Manager {
	m := &vision.Manager{}
	// Use the same approach as vision_test.go
	return m
}

func TestExecutePipeline(t *testing.T) {
	img := createTestImage(100, 100)

	tests := []struct {
		name     string
		steps    []PipelineStep
		tools    *Tools
		wantImg  bool
		wantText bool
		wantErr  bool
	}{
		{
			name: "capture screen and return image",
			steps: []PipelineStep{
				{Tool: "capture_screen", Parameters: map[string]any{}},
			},
			tools: &Tools{
				capture: &mockCapture{
					images: map[string]image.Image{"": img},
				},
			},
			wantImg:  true,
			wantText: false,
			wantErr:  false,
		},
		{
			name: "capture window and return image",
			steps: []PipelineStep{
				{Tool: "capture_window", Parameters: map[string]any{"title": "Terminal"}},
			},
			tools: &Tools{
				capture: &mockCapture{
					images: map[string]image.Image{"window:Terminal": img},
				},
			},
			wantImg:  true,
			wantText: false,
			wantErr:  false,
		},
		{
			name: "capture region with explicit coords",
			steps: []PipelineStep{
				{Tool: "capture_region", Parameters: map[string]any{
					"x":      float64(0),
					"y":      float64(0),
					"width":  float64(100),
					"height": float64(100),
				}},
			},
			tools: &Tools{
				capture: &mockCapture{
					images: map[string]image.Image{"region": img},
				},
			},
			wantImg:  true,
			wantText: false,
			wantErr:  false,
		},
		{
			name: "unknown tool",
			steps: []PipelineStep{
				{Tool: "invalid_tool", Parameters: map[string]any{}},
			},
			tools:   NewTools(&mockCapture{}),
			wantErr: true,
		},
		{
			name:    "empty pipeline",
			steps:   []PipelineStep{},
			tools:   NewTools(&mockCapture{}),
			wantErr: true,
		},
		{
			name: "wait_for step",
			steps: []PipelineStep{
				{Tool: "wait_for", Parameters: map[string]any{"seconds": float64(0)}},
				{Tool: "capture_screen", Parameters: map[string]any{}},
			},
			tools: &Tools{
				capture: &mockCapture{
					images: map[string]image.Image{"": img},
				},
			},
			wantImg:  true,
			wantText: false,
			wantErr:  false,
		},
		{
			name: "wait_for invalid seconds too high",
			steps: []PipelineStep{
				{Tool: "wait_for", Parameters: map[string]any{"seconds": float64(31)}},
			},
			tools:   NewTools(&mockCapture{}),
			wantErr: true,
		},
		{
			name: "wait_for negative seconds",
			steps: []PipelineStep{
				{Tool: "wait_for", Parameters: map[string]any{"seconds": float64(-1)}},
			},
			tools:   NewTools(&mockCapture{}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imgBase64, text, err := ExecutePipeline(context.Background(), tt.steps, tt.tools)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecutePipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.wantImg && imgBase64 == "" {
					t.Error("ExecutePipeline() expected image output, got empty")
				}
				if tt.wantText && text == "" {
					t.Error("ExecutePipeline() expected text output, got empty")
				}
				if !tt.wantImg && imgBase64 != "" {
					t.Error("ExecutePipeline() unexpected image output")
				}
				if !tt.wantText && text != "" {
					t.Error("ExecutePipeline() unexpected text output")
				}
			}
		})
	}
}

func TestPipelineStack(t *testing.T) {
	t.Run("push and pop", func(t *testing.T) {
		stack := &pipelineStack{}
		stack.push("test")
		item, err := stack.pop()
		if err != nil {
			t.Errorf("pop() unexpected error = %v", err)
		}
		if item != "test" {
			t.Errorf("pop() = %v, want %v", item, "test")
		}
	})

	t.Run("pop empty stack", func(t *testing.T) {
		stack := &pipelineStack{}
		_, err := stack.pop()
		if err == nil {
			t.Error("pop() on empty stack should return error")
		}
	})

	t.Run("popImage with valid image", func(t *testing.T) {
		stack := &pipelineStack{}
		imgData := []byte("imagedata")
		stack.push(imgData)
		result, err := stack.popImage()
		if err != nil {
			t.Errorf("popImage() unexpected error = %v", err)
		}
		if string(result) != "imagedata" {
			t.Errorf("popImage() = %v, want %v", string(result), "imagedata")
		}
	})

	t.Run("popImage with invalid type", func(t *testing.T) {
		stack := &pipelineStack{}
		stack.push("not an image")
		_, err := stack.popImage()
		if err == nil {
			t.Error("popImage() with invalid type should return error")
		}
	})

	t.Run("popImage empty stack", func(t *testing.T) {
		stack := &pipelineStack{}
		_, err := stack.popImage()
		if err == nil {
			t.Error("popImage() on empty stack should return error")
		}
	})

	t.Run("popRegion from map", func(t *testing.T) {
		stack := &pipelineStack{}
		stack.push(map[string]any{
			"x":      float64(10),
			"y":      float64(20),
			"width":  float64(100),
			"height": float64(200),
		})
		x, y, w, h, err := stack.popRegion()
		if err != nil {
			t.Errorf("popRegion() unexpected error = %v", err)
		}
		if x != 10 || y != 20 || w != 100 || h != 200 {
			t.Errorf("popRegion() = (%v, %v, %v, %v), want (10, 20, 100, 200)", x, y, w, h)
		}
	})

	t.Run("popRegion invalid type", func(t *testing.T) {
		stack := &pipelineStack{}
		stack.push(12345)
		_, _, _, _, err := stack.popRegion()
		if err == nil {
			t.Error("popRegion() with invalid type should return error")
		}
	})

	t.Run("popRegion empty stack", func(t *testing.T) {
		stack := &pipelineStack{}
		_, _, _, _, err := stack.popRegion()
		if err == nil {
			t.Error("popRegion() on empty stack should return error")
		}
	})
}

func TestExecCaptureRegionFromStack(t *testing.T) {
	img := createTestImage(100, 100)

	tools := &Tools{
		capture: &mockCapture{
			images: map[string]image.Image{"region": img},
		},
	}
	stack := &pipelineStack{}

	// Push a region onto the stack
	stack.push(map[string]any{
		"x":      float64(0),
		"y":      float64(0),
		"width":  float64(50),
		"height": float64(50),
	})

	err := execCaptureRegion(context.Background(), map[string]any{}, tools, stack)
	if err != nil {
		t.Errorf("execCaptureRegion() from stack error = %v", err)
	}

	// Stack should have an image now
	if len(stack.items) != 1 {
		t.Errorf("execCaptureRegion() should push 1 item, got %v", len(stack.items))
	}
}

func TestExecFindRegionMissingImage(t *testing.T) {
	tools := &Tools{}
	stack := &pipelineStack{}

	err := execFindRegion(context.Background(), map[string]any{"description": "test"}, tools, stack)
	if err == nil {
		t.Error("execFindRegion() with empty stack should return error")
	}
}

func TestExecExtractTextMissingImage(t *testing.T) {
	tools := &Tools{}
	stack := &pipelineStack{}

	err := execExtractText(context.Background(), map[string]any{}, tools, stack)
	if err == nil {
		t.Error("execExtractText() with empty stack should return error")
	}
}

func TestExecAnalyzeImageMissingPrompt(t *testing.T) {
	tools := &Tools{}
	stack := &pipelineStack{}
	stack.push([]byte("imagedata"))

	err := execAnalyzeImage(context.Background(), map[string]any{}, tools, stack)
	if err == nil {
		t.Error("execAnalyzeImage() without prompt should return error")
	}
}

func TestExecCompareImagesMissingImages(t *testing.T) {
	tools := &Tools{}
	stack := &pipelineStack{}

	// Only push one image
	stack.push([]byte("image1"))

	err := execCompareImages(context.Background(), map[string]any{}, tools, stack)
	if err == nil {
		t.Error("execCompareImages() with only one image should return error")
	}
}

func TestExecWaitForInvalidSeconds(t *testing.T) {
	err := execWaitFor(map[string]any{"seconds": "not_a_number"})
	if err == nil {
		t.Error("execWaitFor() with invalid seconds should return error")
	}
}

func TestExecutePipelineCaptureRegionMissingCoordsAndStack(t *testing.T) {
	tools := &Tools{
		capture: &mockCapture{
			images: map[string]image.Image{"region": createTestImage(100, 100)},
		},
	}
	stack := &pipelineStack{}

	err := execCaptureRegion(context.Background(), map[string]any{}, tools, stack)
	if err == nil {
		t.Error("execCaptureRegion() without coords and empty stack should return error")
	}
}
