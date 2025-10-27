package axonflow

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	config := AxonFlowConfig{
		AgentURL:     "https://test.example.com",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}

	client := NewClient(config)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	// Verify defaults are set
	if client.config.Mode != "production" {
		t.Errorf("Expected default mode 'production', got '%s'", client.config.Mode)
	}

	if client.config.Timeout != 60*time.Second {
		t.Errorf("Expected default timeout 60s, got %v", client.config.Timeout)
	}

	if !client.config.Retry.Enabled {
		t.Error("Expected retry to be enabled by default")
	}

	if client.config.Retry.MaxAttempts != 3 {
		t.Errorf("Expected default retry attempts 3, got %d", client.config.Retry.MaxAttempts)
	}

	if !client.config.Cache.Enabled {
		t.Error("Expected cache to be enabled by default")
	}

	if client.config.Cache.TTL != 60*time.Second {
		t.Errorf("Expected default cache TTL 60s, got %v", client.config.Cache.TTL)
	}
}

func TestNewClientSimple(t *testing.T) {
	client := NewClientSimple("https://test.example.com", "client-id", "secret")

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.config.AgentURL != "https://test.example.com" {
		t.Errorf("Expected AgentURL 'https://test.example.com', got '%s'", client.config.AgentURL)
	}

	if client.config.ClientID != "client-id" {
		t.Errorf("Expected ClientID 'client-id', got '%s'", client.config.ClientID)
	}
}

func TestSandbox(t *testing.T) {
	client := Sandbox("test-key")

	if client == nil {
		t.Fatal("Expected sandbox client to be created, got nil")
	}

	if client.config.Mode != "sandbox" {
		t.Errorf("Expected sandbox mode, got '%s'", client.config.Mode)
	}

	if !client.config.Debug {
		t.Error("Expected debug mode to be enabled in sandbox")
	}

	if client.config.ClientID != "test-key" {
		t.Errorf("Expected ClientID 'test-key', got '%s'", client.config.ClientID)
	}
}

func TestSandboxDefaultKey(t *testing.T) {
	client := Sandbox("")

	if client == nil {
		t.Fatal("Expected sandbox client to be created, got nil")
	}

	if client.config.ClientID != "demo-key" {
		t.Errorf("Expected default ClientID 'demo-key', got '%s'", client.config.ClientID)
	}
}

func TestCacheBasicOperations(t *testing.T) {
	cache := newCache(1 * time.Second)

	// Test set and get
	cache.set("key1", "value1")
	value, found := cache.get("key1")

	if !found {
		t.Error("Expected to find cached value")
	}

	if value != "value1" {
		t.Errorf("Expected value 'value1', got '%v'", value)
	}

	// Test non-existent key
	_, found = cache.get("nonexistent")
	if found {
		t.Error("Expected key not to be found")
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := newCache(50 * time.Millisecond)

	cache.set("key1", "value1")

	// Value should exist immediately
	_, found := cache.get("key1")
	if !found {
		t.Error("Expected cached value to exist")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Value should be expired
	_, found = cache.get("key1")
	if found {
		t.Error("Expected cached value to be expired")
	}
}

func TestMinHelper(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{10, 10, 10},
		{0, 100, 0},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestHttpError(t *testing.T) {
	err := &httpError{
		statusCode: 404,
		message:    "Not Found",
	}

	expected := "HTTP 404: Not Found"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestClientResponseStruct(t *testing.T) {
	// Test that ClientResponse can be created with all fields
	resp := ClientResponse{
		Success:     true,
		Data:        "test data",
		Result:      "test result",
		PlanID:      "plan-123",
		RequestID:   "req-456",
		Error:       "",
		Blocked:     false,
		BlockReason: "",
		PolicyInfo: &PolicyEvaluationInfo{
			PoliciesEvaluated: []string{"policy1", "policy2"},
			StaticChecks:      []string{"check1"},
			ProcessingTime:    150,
			TenantID:          "tenant-1",
		},
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}

	if resp.PolicyInfo.ProcessingTime != 150 {
		t.Errorf("Expected ProcessingTime 150, got %d", resp.PolicyInfo.ProcessingTime)
	}

	if len(resp.PolicyInfo.PoliciesEvaluated) != 2 {
		t.Errorf("Expected 2 policies evaluated, got %d", len(resp.PolicyInfo.PoliciesEvaluated))
	}
}

func TestPlanStepStruct(t *testing.T) {
	// Test that PlanStep has correct field names
	step := PlanStep{
		ID:            "step-1",
		Name:          "Test Step",
		Type:          "query",
		Description:   "Test description",
		Dependencies:  []string{"step-0"},
		Agent:         "test-agent",
		EstimatedTime: "2s",
		Parameters:    map[string]interface{}{"key": "value"},
	}

	if step.ID != "step-1" {
		t.Errorf("Expected ID 'step-1', got '%s'", step.ID)
	}

	if len(step.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(step.Dependencies))
	}

	if step.EstimatedTime != "2s" {
		t.Errorf("Expected EstimatedTime '2s', got '%s'", step.EstimatedTime)
	}
}

func TestPlanExecutionResponseStruct(t *testing.T) {
	// Test that PlanExecutionResponse has all required fields
	resp := PlanExecutionResponse{
		PlanID:         "plan-123",
		Status:         "completed",
		Result:         "Test result",
		CompletedSteps: 5,
		TotalSteps:     5,
		StepResults: []StepResult{
			{
				StepID:   "step-1",
				StepName: "Test Step",
				Status:   "completed",
				Result:   "Success",
				Duration: "1s",
			},
		},
		Duration: "10s",
	}

	if resp.CompletedSteps != 5 {
		t.Errorf("Expected CompletedSteps 5, got %d", resp.CompletedSteps)
	}

	if resp.TotalSteps != 5 {
		t.Errorf("Expected TotalSteps 5, got %d", resp.TotalSteps)
	}

	if len(resp.StepResults) != 1 {
		t.Errorf("Expected 1 step result, got %d", len(resp.StepResults))
	}

	if resp.StepResults[0].StepID != "step-1" {
		t.Errorf("Expected StepID 'step-1', got '%s'", resp.StepResults[0].StepID)
	}
}

func TestConnectorMetadataStruct(t *testing.T) {
	// Test that ConnectorMetadata has InstanceName field
	metadata := ConnectorMetadata{
		ID:           "connector-1",
		Name:         "Test Connector",
		Type:         "http",
		Version:      "1.0.0",
		Description:  "Test description",
		Installed:    true,
		InstanceName: "test-instance",
	}

	if !metadata.Installed {
		t.Error("Expected Installed to be true")
	}

	if metadata.InstanceName != "test-instance" {
		t.Errorf("Expected InstanceName 'test-instance', got '%s'", metadata.InstanceName)
	}
}

func TestConfigurationEdgeCases(t *testing.T) {
	// Test that configuration with zero values gets defaults
	config := AxonFlowConfig{
		AgentURL:     "https://test.example.com",
		ClientID:     "test",
		ClientSecret: "secret",
		Timeout:      0, // Should get default
		Retry: RetryConfig{
			Enabled:      false,
			MaxAttempts:  0, // Should get default
			InitialDelay: 0, // Should get default
		},
		Cache: CacheConfig{
			Enabled: false,
			TTL:     0, // Should get default
		},
	}

	client := NewClient(config)

	// Verify defaults were applied
	if client.config.Timeout != 60*time.Second {
		t.Errorf("Expected default timeout, got %v", client.config.Timeout)
	}

	if client.config.Retry.MaxAttempts != 3 {
		t.Errorf("Expected default retry attempts, got %d", client.config.Retry.MaxAttempts)
	}

	if client.config.Retry.InitialDelay != 1*time.Second {
		t.Errorf("Expected default initial delay, got %v", client.config.Retry.InitialDelay)
	}

	if client.config.Cache.TTL != 60*time.Second {
		t.Errorf("Expected default cache TTL, got %v", client.config.Cache.TTL)
	}
}
