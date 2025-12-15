// sdk_integration_test.go - Integration tests for AxonFlow Go SDK
// Run with: go test -v -tags=integration ./...
//
//go:build integration
// +build integration

package axonflow

import (
	"os"
	"strings"
	"testing"
	"time"
)

// Integration test configuration
// Set these environment variables before running:
//   AXONFLOW_AGENT_URL=http://localhost:8080
//   AXONFLOW_CLIENT_ID=demo-client
//   AXONFLOW_CLIENT_SECRET=demo-secret

func getTestConfig(t *testing.T) AxonFlowConfig {
	agentURL := os.Getenv("AXONFLOW_AGENT_URL")
	if agentURL == "" {
		agentURL = "http://localhost:8080"
	}

	clientID := os.Getenv("AXONFLOW_CLIENT_ID")
	if clientID == "" {
		clientID = "demo-client"
	}

	clientSecret := os.Getenv("AXONFLOW_CLIENT_SECRET")
	if clientSecret == "" {
		clientSecret = "demo-secret"
	}

	return AxonFlowConfig{
		AgentURL:     agentURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Debug:        true,
		Timeout:      30 * time.Second,
	}
}

// TestHealthCheck verifies basic connectivity
func TestIntegration_HealthCheck(t *testing.T) {
	client := NewClient(getTestConfig(t))

	err := client.HealthCheck()
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	t.Log("Health check passed")
}

// TestExecuteQuery_Simple tests a basic query
func TestIntegration_ExecuteQuery_Simple(t *testing.T) {
	client := NewClient(getTestConfig(t))

	resp, err := client.ExecuteQuery("demo-user", "What is 2+2?", "chat", nil)
	if err != nil {
		t.Fatalf("ExecuteQuery failed: %v", err)
	}

	if !resp.Success && !resp.Blocked {
		t.Errorf("Expected success or blocked, got error: %s", resp.Error)
	}
	t.Logf("Response: success=%v, blocked=%v", resp.Success, resp.Blocked)
}

// TestExecuteQuery_SQLInjection tests that SQL injection is blocked
func TestIntegration_ExecuteQuery_SQLInjection(t *testing.T) {
	client := NewClient(getTestConfig(t))

	// SQL injection should be blocked
	resp, err := client.ExecuteQuery("demo-user", "SELECT * FROM users; DROP TABLE users;--", "sql", nil)
	if err != nil {
		t.Fatalf("ExecuteQuery failed: %v", err)
	}

	if !resp.Blocked {
		t.Errorf("Expected SQL injection to be blocked, got blocked=%v", resp.Blocked)
	}
	t.Logf("SQL injection blocked: %s", resp.BlockReason)
}

// TestExecuteQuery_PIIDetection tests that PII is blocked
func TestIntegration_ExecuteQuery_PIIDetection(t *testing.T) {
	client := NewClient(getTestConfig(t))

	// SSN should be blocked (with PII_BLOCK_CRITICAL=true default)
	resp, err := client.ExecuteQuery("demo-user", "My SSN is 123-45-6789", "chat", nil)
	if err != nil {
		t.Fatalf("ExecuteQuery failed: %v", err)
	}

	if !resp.Blocked {
		t.Errorf("Expected SSN to be blocked, got blocked=%v", resp.Blocked)
	}
	t.Logf("PII blocked: %s", resp.BlockReason)
}

// TestGatewayMode_PreCheck tests Gateway Mode pre-check
func TestIntegration_GatewayMode_PreCheck(t *testing.T) {
	client := NewClient(getTestConfig(t))

	result, err := client.GetPolicyApprovedContext("demo-user", "Analyze this data", nil, nil)
	if err != nil {
		t.Fatalf("GetPolicyApprovedContext failed: %v", err)
	}

	if result.ContextID == "" {
		t.Error("Expected non-empty context_id")
	}

	// Check that ExpiresAt was parsed correctly (should be in the future)
	if result.ExpiresAt.IsZero() {
		t.Error("ExpiresAt was not parsed correctly (zero value)")
	}
	if result.ExpiresAt.Before(time.Now()) {
		t.Errorf("ExpiresAt should be in the future, got %v", result.ExpiresAt)
	}

	t.Logf("Pre-check: context_id=%s, approved=%v, expires_at=%v",
		result.ContextID, result.Approved, result.ExpiresAt)
}

// TestGatewayMode_PreCheckWithNanoseconds tests datetime parsing with nanoseconds
func TestIntegration_GatewayMode_PreCheckDatetimeParsing(t *testing.T) {
	client := NewClient(getTestConfig(t))

	result, err := client.GetPolicyApprovedContext("demo-user", "Test datetime parsing", nil, nil)
	if err != nil {
		t.Fatalf("GetPolicyApprovedContext failed: %v", err)
	}

	// ExpiresAt should be approximately 5 minutes from now (default context expiry)
	expectedExpiry := time.Now().Add(5 * time.Minute)
	timeDiff := result.ExpiresAt.Sub(expectedExpiry)

	// Allow 30 second tolerance
	if timeDiff.Abs() > 30*time.Second {
		t.Errorf("ExpiresAt is not within expected range. Got %v, expected ~%v",
			result.ExpiresAt, expectedExpiry)
	}

	t.Logf("Datetime parsing OK: expires_at=%v (diff from expected: %v)",
		result.ExpiresAt, timeDiff)
}

// TestGatewayMode_AuditLLMCall tests Gateway Mode audit
func TestIntegration_GatewayMode_AuditLLMCall(t *testing.T) {
	client := NewClient(getTestConfig(t))

	// First get a context
	preCheck, err := client.GetPolicyApprovedContext("demo-user", "Test audit", nil, nil)
	if err != nil {
		t.Fatalf("GetPolicyApprovedContext failed: %v", err)
	}

	// Then audit an LLM call
	result, err := client.AuditLLMCall(
		preCheck.ContextID,
		"Test response summary",
		"openai",
		"gpt-4",
		TokenUsage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150},
		250,
		nil,
	)
	if err != nil {
		t.Fatalf("AuditLLMCall failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected audit to succeed")
	}
	if result.AuditID == "" {
		t.Error("Expected non-empty audit_id")
	}

	t.Logf("Audit logged: audit_id=%s", result.AuditID)
}

// TestGeneratePlan tests multi-agent plan generation
func TestIntegration_GeneratePlan(t *testing.T) {
	client := NewClient(getTestConfig(t))

	// Test backward-compatible call (no userToken - uses variadic)
	plan, err := client.GeneratePlan("Book a flight from NYC to LA", "travel")
	if err != nil {
		// Plan generation may fail if orchestrator doesn't have LLM configured
		// This is acceptable - we're testing the SDK request format
		if strings.Contains(err.Error(), "LLM") || strings.Contains(err.Error(), "provider") {
			t.Skipf("Plan generation skipped (LLM not configured): %v", err)
		}
		t.Fatalf("GeneratePlan failed: %v", err)
	}

	if plan.PlanID == "" {
		t.Error("Expected non-empty plan_id")
	}

	t.Logf("Plan generated: plan_id=%s, steps=%d", plan.PlanID, len(plan.Steps))
}

// TestGeneratePlanWithUserToken tests plan generation with explicit user token
func TestIntegration_GeneratePlanWithUserToken(t *testing.T) {
	client := NewClient(getTestConfig(t))

	// Test with explicit userToken (variadic parameter)
	plan, err := client.GeneratePlan("Simple query", "generic", "custom-user-token")
	if err != nil {
		if strings.Contains(err.Error(), "LLM") || strings.Contains(err.Error(), "provider") {
			t.Skipf("Plan generation skipped (LLM not configured): %v", err)
		}
		t.Fatalf("GeneratePlan with userToken failed: %v", err)
	}

	if plan.PlanID == "" {
		t.Error("Expected non-empty plan_id")
	}

	t.Logf("Plan with custom token generated: plan_id=%s", plan.PlanID)
}

// TestListConnectors tests listing MCP connectors
func TestIntegration_ListConnectors(t *testing.T) {
	client := NewClient(getTestConfig(t))

	connectors, err := client.ListConnectors()
	if err != nil {
		t.Fatalf("ListConnectors failed: %v", err)
	}

	t.Logf("Found %d connectors", len(connectors))
	for _, c := range connectors {
		t.Logf("  - %s (%s)", c.Name, c.Type)
	}
}
