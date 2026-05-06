package access

import (
	"testing"
	"time"
)

func TestIsExempt(t *testing.T) {
	am := NewAccessManager(nil, 300)
	// Register tools as the real code does
	am.RegisterTool("list_monitors", true)
	am.RegisterTool("list_vision_providers", true)
	am.RegisterTool("get_skill_info_for_agent", true)
	am.RegisterTool("list_tool_access", true)
	am.RegisterTool("allow_tool_access", true)
	am.RegisterTool("list_windows", false)
	am.RegisterTool("capture_screen", false)
	am.RegisterTool("capture_window", false)
	am.RegisterTool("capture_region", false)
	am.RegisterTool("analyze_image", false)
	am.RegisterTool("extract_text", false)
	am.RegisterTool("find_region", false)
	am.RegisterTool("compare_images", false)
	am.RegisterTool("execute_capture_pipeline", false)

	tests := []struct {
		tool   string
		exempt bool
	}{
		{"list_monitors", true},
		{"list_vision_providers", true},
		{"get_skill_info_for_agent", true},
		{"list_tool_access", true},
		{"allow_tool_access", true},
		{"capture_screen", false},
		{"capture_window", false},
		{"capture_region", false},
		{"analyze_image", false},
	}
	for _, tt := range tests {
		if got := am.isExempt(tt.tool); got != tt.exempt {
			t.Errorf("isExempt(%q) = %v, want %v", tt.tool, got, tt.exempt)
		}
	}
}

func TestNewAccessManager(t *testing.T) {
	am := NewAccessManager(nil, 0)
	if am == nil {
		t.Fatal("expected non-nil AccessManager")
	}
	if am.defaultDuration != 300*time.Second {
		t.Errorf("defaultDuration = %v, want 300s", am.defaultDuration)
	}

	am = NewAccessManager(map[string]string{"capture_screen": "deny"}, 60)
	if am.policies["capture_screen"] != "deny" {
		t.Error("policy not set correctly")
	}
	if am.defaultDuration != 60*time.Second {
		t.Errorf("defaultDuration = %v, want 60s", am.defaultDuration)
	}
}

func TestGetAccess_ExemptTools(t *testing.T) {
	am := NewAccessManager(nil, 300)
	// Register exempt tools
	am.RegisterTool("list_monitors", true)
	am.RegisterTool("list_vision_providers", true)
	am.RegisterTool("get_skill_info_for_agent", true)
	am.RegisterTool("list_tool_access", true)
	am.RegisterTool("allow_tool_access", true)

	exempt := []string{"list_monitors", "list_vision_providers", "get_skill_info_for_agent", "list_tool_access", "allow_tool_access"}
	for _, tool := range exempt {
		if got := am.GetAccess(tool); got != "allow" {
			t.Errorf("exempt tool %q: got %q, want %q", tool, got, "allow")
		}
	}
}

func TestGetAccess_ConfiguredPolicies(t *testing.T) {
	am := NewAccessManager(map[string]string{
		"capture_screen": "deny",
		"capture_window": "allow",
		"capture_region": "ask",
	}, 300)

	tests := []struct {
		tool   string
		policy string
	}{
		{"capture_screen", "deny"},
		{"capture_window", "allow"},
		{"capture_region", "ask"},
		{"analyze_image", "ask"}, // not configured, default
	}
	for _, tt := range tests {
		if got := am.GetAccess(tt.tool); got != tt.policy {
			t.Errorf("GetAccess(%q) = %q, want %q", tt.tool, got, tt.policy)
		}
	}
}

func TestGetAccess_InvalidPolicy(t *testing.T) {
	am := NewAccessManager(map[string]string{
		"capture_screen": "invalid",
	}, 300)
	// Invalid policy should fall through to default "ask"
	if got := am.GetAccess("capture_screen"); got != "ask" {
		t.Errorf("invalid policy: got %q, want %q", got, "ask")
	}
}

func TestGetAccess_TemporaryGrant(t *testing.T) {
	am := NewAccessManager(map[string]string{
		"capture_screen": "ask",
	}, 300)

	// Before grant
	if got := am.GetAccess("capture_screen"); got != "ask" {
		t.Errorf("before grant: got %q, want %q", got, "ask")
	}

	// Grant for 1 second
	am.GrantTemporary("capture_screen", 1*time.Second)

	// After grant, should be allowed
	if got := am.GetAccess("capture_screen"); got != "allow" {
		t.Errorf("after grant: got %q, want %q", got, "allow")
	}

	// Wait for expiry
	time.Sleep(1100 * time.Millisecond)

	// After expiry, should be back to "ask"
	if got := am.GetAccess("capture_screen"); got != "ask" {
		t.Errorf("after expiry: got %q, want %q", got, "ask")
	}
}

func TestGetAccess_TemporaryGrant_DefaultDuration(t *testing.T) {
	am := NewAccessManager(map[string]string{
		"capture_screen": "ask",
	}, 1) // 1 second default

	// Grant with 0 duration (use default)
	am.GrantTemporary("capture_screen", 0)

	if got := am.GetAccess("capture_screen"); got != "allow" {
		t.Errorf("after grant: got %q, want %q", got, "allow")
	}

	time.Sleep(1100 * time.Millisecond)

	if got := am.GetAccess("capture_screen"); got != "ask" {
		t.Errorf("after expiry: got %q, want %q", got, "ask")
	}
}

func TestListAccess(t *testing.T) {
	am := NewAccessManager(map[string]string{
		"capture_screen": "deny",
		"capture_window": "allow",
	}, 300)

	// Register all tools as the real code does
	am.RegisterTool("list_monitors", true)
	am.RegisterTool("list_vision_providers", true)
	am.RegisterTool("get_skill_info_for_agent", true)
	am.RegisterTool("list_tool_access", true)
	am.RegisterTool("allow_tool_access", true)
	am.RegisterTool("list_windows", false)
	am.RegisterTool("capture_screen", false)
	am.RegisterTool("capture_window", false)
	am.RegisterTool("capture_region", false)
	am.RegisterTool("analyze_image", false)
	am.RegisterTool("extract_text", false)
	am.RegisterTool("find_region", false)
	am.RegisterTool("compare_images", false)
	am.RegisterTool("execute_capture_pipeline", false)

	access := am.ListAccess()

	// Exempt tools should be "allow"
	if access["list_monitors"] != "allow" {
		t.Error("list_monitors should be allow")
	}

	// Configured tools
	if access["capture_screen"] != "deny" {
		t.Error("capture_screen should be deny")
	}
	if access["capture_window"] != "allow" {
		t.Error("capture_window should be allow")
	}

	// Unconfigured tools default to "ask"
	if access["analyze_image"] != "ask" {
		t.Error("analyze_image should be ask")
	}
}

func TestGrantTemporary_Overwrites(t *testing.T) {
	am := NewAccessManager(map[string]string{
		"capture_screen": "ask",
	}, 300)

	// Grant for 5 seconds
	am.GrantTemporary("capture_screen", 5*time.Second)
	if got := am.GetAccess("capture_screen"); got != "allow" {
		t.Errorf("after first grant: got %q, want %q", got, "allow")
	}

	// Grant again for 1 second - should overwrite
	am.GrantTemporary("capture_screen", 1*time.Second)
	time.Sleep(1100 * time.Millisecond)

	if got := am.GetAccess("capture_screen"); got != "ask" {
		t.Errorf("after second grant expiry: got %q, want %q", got, "ask")
	}
}
