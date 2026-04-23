// Copyright 2025 Emmanuel Deloget. All rights reserved.
// Use of this source code is governed by the license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      any           `json:"id"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type MCPResponse struct {
	Content []map[string]any `json:"content"`
	IsError bool             `json:"isError"`
}

func main() {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:11777"
	}

	outputDir := os.Getenv("OUTPUT_DIR")
	if outputDir == "" {
		outputDir = "/tmp/screenshooter-mcp-images"
	}

	if len(os.Args) > 1 {
		serverURL = os.Args[1]
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	ctx := context.Background()

	monitors, err := callListMonitors(ctx, client, serverURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list_monitors failed: %v\n", err)
		os.Exit(1)
	}

	saveJSON(ctx, outputDir, "list_monitors.json", monitors)

	windows, err := callListWindows(ctx, client, serverURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list_windows failed: %v\n", err)
		os.Exit(1)
	}

	saveJSON(ctx, outputDir, "list_windows.json", windows)

	imgData, err := callCaptureScreen(ctx, client, serverURL, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "capture_screen failed: %v\n", err)
		os.Exit(1)
	}

	saveImage(ctx, outputDir, "capture_screen.png", imgData)

	if len(monitors) > 0 {
		firstMonitor := monitors[0]
		if name, ok := firstMonitor["Name"].(string); ok {
			imgData, err := callCaptureScreen(ctx, client, serverURL, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "capture_screen (monitor %s) failed: %v\n", name, err)
				os.Exit(1)
			}
			saveImage(ctx, outputDir, "capture_screen-"+name+".png", imgData)
		}
	}

	if len(windows) > 0 {
		firstWindow := windows[0]
		if title, ok := firstWindow["Title"].(string); ok {
			imgData, err := callCaptureWindow(ctx, client, serverURL, title)
			if err != nil {
				fmt.Fprintf(os.Stderr, "capture_window (title %s) failed: %v\n", title, err)
				os.Exit(1)
			}
			saveImage(ctx, outputDir, "capture_window-"+title+".png", imgData)
		}
	}

	imgData, err = callCaptureRegion(ctx, client, serverURL, 0, 0, 800, 600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "capture_region failed: %v\n", err)
		os.Exit(1)
	}
	saveImage(ctx, outputDir, "capture_region.png", imgData)

	fmt.Println("All tests passed!")
}

func callListMonitors(ctx context.Context, client *http.Client, serverURL string) ([]map[string]any, error) {
	params := ToolCallParams{
		Name: "list_monitors",
	}
	return callTool(ctx, client, serverURL, params)
}

func callListWindows(ctx context.Context, client *http.Client, serverURL string) ([]map[string]any, error) {
	params := ToolCallParams{
		Name: "list_windows",
	}
	return callTool(ctx, client, serverURL, params)
}

func callCaptureScreen(ctx context.Context, client *http.Client, serverURL, monitor string) ([]byte, error) {
	args := map[string]any{}
	if monitor != "" {
		args["monitor"] = monitor
	}

	params := ToolCallParams{
		Name:      "capture_screen",
		Arguments: mustMarshal(args),
	}

	result, err := callToolRaw(ctx, client, serverURL, params)
	if err != nil {
		return nil, err
	}

	return extractImage(result)
}

func callCaptureWindow(ctx context.Context, client *http.Client, serverURL, title string) ([]byte, error) {
	args := map[string]any{
		"title": title,
	}

	params := ToolCallParams{
		Name:      "capture_window",
		Arguments: mustMarshal(args),
	}

	result, err := callToolRaw(ctx, client, serverURL, params)
	if err != nil {
		return nil, err
	}

	return extractImage(result)
}

func callCaptureRegion(ctx context.Context, client *http.Client, serverURL string, x, y, w, h int) ([]byte, error) {
	args := map[string]any{
		"x":      x,
		"y":      y,
		"width":  w,
		"height": h,
	}

	params := ToolCallParams{
		Name:      "capture_region",
		Arguments: mustMarshal(args),
	}

	result, err := callToolRaw(ctx, client, serverURL, params)
	if err != nil {
		return nil, err
	}

	return extractImage(result)
}

func callTool(ctx context.Context, client *http.Client, serverURL string, params ToolCallParams) ([]map[string]any, error) {
	result, err := callToolRaw(ctx, client, serverURL, params)
	if err != nil {
		return nil, err
	}

	data, ok := result["content"].([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	if len(data) == 0 {
		return []map[string]any{}, nil
	}

	first, ok := data[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected content format")
	}

	text, ok := first["text"].(string)
	if !ok {
		return nil, fmt.Errorf("expected text content")
	}

	var parsed []map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return parsed, nil
}

func callToolRaw(ctx context.Context, client *http.Client, serverURL string, params ToolCallParams) (map[string]any, error) {
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  mustMarshal(params),
		ID:      1,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", serverURL, io.NopCloser(
		io.NopCloser(
			&readCloseProxy{data: body},
		),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %d - %s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
	}

	result, ok := rpcResp.Result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	return result, nil
}

func extractImage(result map[string]any) ([]byte, error) {
	content, ok := result["content"].([]any)
	if !ok {
		return nil, fmt.Errorf("no content in result")
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("empty content")
	}

	first, ok := content[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected content type")
	}

	data, ok := first["data"].([]any)
	if !ok {
		return nil, fmt.Errorf("no data field")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty data array")
	}

	dataStr, ok := data[0].(string)
	if !ok {
		return nil, fmt.Errorf("expected string data")
	}

	return []byte(dataStr), nil
}

func mustMarshal(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func saveJSON(ctx context.Context, dir, name string, data any) {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal %s: %v\n", name, err)
		return
	}

	path := dir + "/" + name
	if err := os.WriteFile(path, content, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", path, err)
		return
	}

	fmt.Printf("Saved: %s\n", path)
}

func saveImage(ctx context.Context, dir, name string, data []byte) {
	path := dir + "/" + name
	if err := os.WriteFile(path, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", path, err)
		return
	}

	fmt.Printf("Saved: %s (%d bytes)\n", path, len(data))
}

type readCloseProxy struct {
	data []byte
	pos  int
}

func (r *readCloseProxy) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *readCloseProxy) Close() error {
	return nil
}
