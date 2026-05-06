package access

import (
	"sync"
	"time"
)

// AccessManager controls tool access based on configured policies.
type AccessManager struct {
	policies        map[string]string
	tempGrants      map[string]time.Time
	defaultDuration time.Duration
	mu              sync.Mutex
	tools           map[string]bool // tool name -> exempt status
}

// RegisterTool registers a tool with the access manager.
func (am *AccessManager) RegisterTool(name string, exempt bool) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.tools[name] = exempt
}

// isExempt returns true if the tool is always allowed.
func (am *AccessManager) isExempt(tool string) bool {
	am.mu.Lock()
	defer am.mu.Unlock()
	exempt, ok := am.tools[tool]
	return ok && exempt
}

// NewAccessManager creates an AccessManager from config.
// policies is the access map from config.json.
// defaultDuration is the default temporary grant duration in seconds.
func NewAccessManager(policies map[string]string, defaultDuration int) *AccessManager {
	if defaultDuration <= 0 {
		defaultDuration = 300
	}
	return &AccessManager{
		policies:        policies,
		tempGrants:      make(map[string]time.Time),
		defaultDuration: time.Duration(defaultDuration) * time.Second,
		tools:           make(map[string]bool),
	}
}

// GetAccess returns the effective access policy for a tool.
// Returns: "allow", "deny", or "ask".
func (am *AccessManager) GetAccess(tool string) string {
	if am.isExempt(tool) {
		return "allow"
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	// Check temporary grants
	if expiry, ok := am.tempGrants[tool]; ok {
		if time.Now().Before(expiry) {
			return "allow"
		}
		delete(am.tempGrants, tool)
	}

	// Check configured policy
	if policy, ok := am.policies[tool]; ok {
		switch policy {
		case "allow", "deny", "ask":
			return policy
		}
	}

	// Default
	return "ask"
}

// GrantTemporary allows a tool for a limited time.
// If duration is 0, uses the default duration.
func (am *AccessManager) GrantTemporary(tool string, duration time.Duration) {
	if duration <= 0 {
		duration = am.defaultDuration
	}
	am.mu.Lock()
	am.tempGrants[tool] = time.Now().Add(duration)
	am.mu.Unlock()
}

// IsToolRegistered returns true if the tool has been registered with the access manager.
func (am *AccessManager) IsToolRegistered(tool string) bool {
	am.mu.Lock()
	defer am.mu.Unlock()
	_, ok := am.tools[tool]
	return ok
}

// ListAccess returns all tools with their effective access status.
func (am *AccessManager) ListAccess() map[string]string {
	am.mu.Lock()
	tools := make([]string, 0, len(am.tools))
	for name := range am.tools {
		tools = append(tools, name)
	}
	am.mu.Unlock()

	result := make(map[string]string, len(tools))
	for _, name := range tools {
		result[name] = am.GetAccess(name)
	}
	return result
}
