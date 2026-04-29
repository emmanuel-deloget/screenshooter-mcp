// Copyright 2025 Emmanuel Deloget. All rights reserved.
// Use of this source code is governed by the license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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

type InitializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      ClientInfo     `json:"clientInfo"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ServerInfo      ServerInfo     `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type MCPServer struct {
	client      *http.Client
	serverURL   string
	sessionID   string
	initialized bool
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

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	mcp := &MCPServer{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		serverURL: serverURL,
	}

	if err := mcp.initialize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "initialize failed: %v\n", err)
		os.Exit(1)
	}

	allPassed := true
	var monitors []map[string]any
	var windows []map[string]any

	if m, err := mcp.listMonitors(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "list_monitors failed: %v\n", err)
		allPassed = false
	} else {
		monitors = m
		saveJSON(ctx, outputDir, "list_monitors.json", monitors)
	}

	if w, err := mcp.listWindows(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "list_windows failed: %v\n", err)
		allPassed = false
	} else {
		windows = w
		saveJSON(ctx, outputDir, "list_windows.json", windows)
	}

	if imgData, err := mcp.captureScreen(ctx, ""); err != nil {
		fmt.Fprintf(os.Stderr, "capture_screen failed: %v\n", err)
		allPassed = false
	} else {
		saveImage(ctx, outputDir, "capture_screen.png", imgData)
	}

	if len(monitors) > 0 {
		firstMonitor := monitors[0]
		if name, ok := firstMonitor["Name"].(string); ok {
			if imgData, err := mcp.captureScreen(ctx, name); err != nil {
				fmt.Fprintf(os.Stderr, "capture_screen (monitor %s) failed: %v\n", name, err)
				allPassed = false
			} else {
				saveImage(ctx, outputDir, "capture_screen-"+name+".png", imgData)
			}
		}
	}

	if len(windows) > 0 {
		firstWindow := windows[0]
		if title, ok := firstWindow["Title"].(string); ok {
			if imgData, err := mcp.captureWindow(ctx, title); err != nil {
				fmt.Fprintf(os.Stderr, "capture_window (title %s) failed: %v\n", title, err)
				allPassed = false
			} else {
				saveImage(ctx, outputDir, "capture_window-"+title+".png", imgData)
			}
		}
	}

	if imgData, err := mcp.captureRegion(ctx, 0, 0, 800, 600); err != nil {
		fmt.Fprintf(os.Stderr, "capture_region failed: %v\n", err)
		allPassed = false
	} else {
		saveImage(ctx, outputDir, "capture_region.png", imgData)
	}

	if allPassed {
		fmt.Println("All tests passed!")
	} else {
		fmt.Println("Some tests failed!")
		os.Exit(1)
	}
}

func (m *MCPServer) initialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    map[string]any{},
		ClientInfo: ClientInfo{
			Name:    "test-mcp",
			Version: "1.0.0",
		},
	}

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		Params:  mustMarshal(params),
		ID:      1,
	}

	_, err := m.call(ctx, reqBody)
	if err != nil {
		return fmt.Errorf("initialize failed: %w", err)
	}

	m.initialized = true

	notifReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
		ID:      2,
	}

	m.call(ctx, notifReq)

	time.Sleep(100 * time.Millisecond)

	fmt.Fprintf(os.Stderr, "Session ID: %s\n", m.sessionID)

	return nil
}

func (m *MCPServer) listMonitors(ctx context.Context) ([]map[string]any, error) {
	params := ToolCallParams{
		Name: "list_monitors",
	}
	return m.callTool(ctx, params)
}

func (m *MCPServer) listWindows(ctx context.Context) ([]map[string]any, error) {
	params := ToolCallParams{
		Name: "list_windows",
	}
	return m.callTool(ctx, params)
}

func (m *MCPServer) captureScreen(ctx context.Context, monitor string) ([]byte, error) {
	args := map[string]any{}
	if monitor != "" {
		args["monitor"] = monitor
	}

	params := ToolCallParams{
		Name:      "capture_screen",
		Arguments: mustMarshal(args),
	}

	result, err := m.callToolRaw(ctx, params)
	if err != nil {
		return nil, err
	}

	return extractImage(result)
}

func (m *MCPServer) captureWindow(ctx context.Context, title string) ([]byte, error) {
	args := map[string]any{
		"title": title,
	}

	params := ToolCallParams{
		Name:      "capture_window",
		Arguments: mustMarshal(args),
	}

	result, err := m.callToolRaw(ctx, params)
	if err != nil {
		return nil, err
	}

	return extractImage(result)
}

func (m *MCPServer) captureRegion(ctx context.Context, x, y, w, h int) ([]byte, error) {
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

	result, err := m.callToolRaw(ctx, params)
	if err != nil {
		return nil, err
	}

	return extractImage(result)
}

func (m *MCPServer) callTool(ctx context.Context, params ToolCallParams) ([]map[string]any, error) {
	result, err := m.callToolRaw(ctx, params)
	if err != nil {
		return nil, err
	}

	if isError, ok := result["isError"].(bool); ok && isError {
		data, ok := result["content"].([]any)
		if ok && len(data) > 0 {
			if first, ok := data[0].(map[string]any); ok {
				if text, ok := first["text"].(string); ok {
					return nil, fmt.Errorf("tool returned error: %s", text)
				}
			}
		}
		return nil, fmt.Errorf("tool returned error (no details)")
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

func (m *MCPServer) callToolRaw(ctx context.Context, params ToolCallParams) (map[string]any, error) {
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params:  mustMarshal(params),
		ID:      3,
	}
	return m.call(ctx, reqBody)
}

func (m *MCPServer) call(ctx context.Context, req JSONRPCRequest) (map[string]any, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.serverURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")

	if m.sessionID != "" {
		httpReq.Header.Set("mcp-session-id", m.sessionID)
	}

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized && !m.initialized {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unauthorized: %s", string(respBody))
	}

	if resp.StatusCode == http.StatusNotFound && m.sessionID != "" {
		return nil, fmt.Errorf("session not found: %s", m.sessionID)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %d - %s", resp.StatusCode, string(respBody))
	}

	sessionID := resp.Header.Get("mcp-session-id")
	if sessionID != "" && m.sessionID == "" {
		m.sessionID = sessionID
		fmt.Fprintf(os.Stderr, "Session ID: %s\n", m.sessionID)
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/event-stream") {
		return parseSSEResponse(resp.Body)
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
	if !ok && rpcResp.Result != nil {
		return nil, fmt.Errorf("unexpected result type: %T", rpcResp.Result)
	}

	return result, nil
}

func parseSSEResponse(body io.Reader) (map[string]any, error) {
	scanner := bufio.NewScanner(body)
	buf := make([]byte, 256*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var dataBuf strings.Builder

	flush := func() (map[string]any, error) {
		if dataBuf.Len() == 0 {
			return nil, nil
		}

		raw := dataBuf.String()
		dataBuf.Reset()

		var rpcResp JSONRPCResponse
		if err := json.Unmarshal([]byte(raw), &rpcResp); err != nil {
			return nil, err
		}

		if rpcResp.Error != nil {
			return nil, fmt.Errorf("RPC error: %s", rpcResp.Error.Message)
		}

		result, ok := rpcResp.Result.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unexpected result type in SSE")
		}

		return result, nil
	}

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "event:"):
			continue

		case strings.HasPrefix(line, "data:"):
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			dataBuf.WriteString(data)

		case line == "":
			res, err := flush()
			if err != nil {
				return nil, err
			}
			if res != nil {
				return res, nil
			}
		}
	}

	// flush final
	return flush()
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

	dataField, ok := first["data"]
	if !ok {
		return nil, fmt.Errorf("no data field")
	}

	var dataBytes []byte
	switch v := dataField.(type) {
	case string:
		var err error
		dataBytes, err = decoding(v)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", err)
		}
	case []any:
		if len(v) == 0 {
			return nil, fmt.Errorf("empty data array")
		}
		dataStr, ok := v[0].(string)
		if !ok {
			return nil, fmt.Errorf("expected string data")
		}
		dataBytes = []byte(dataStr)
	default:
		return nil, fmt.Errorf("unexpected data type: %T", dataField)
	}

	return dataBytes, nil
}

func decoding(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
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
