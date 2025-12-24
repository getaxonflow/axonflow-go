package axonflow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Sample test data
var sampleStaticPolicy = StaticPolicy{
	ID:          "pol_123",
	Name:        "Block SQL Injection",
	Description: "Blocks SQL injection attempts",
	Category:    CategorySecuritySQLI,
	Tier:        TierSystem,
	Pattern:     "(?i)(union\\s+select|drop\\s+table)",
	Severity:    9,
	Enabled:     true,
	Action:      ActionBlock,
	CreatedAt:   time.Now(),
	UpdatedAt:   time.Now(),
	Version:     1,
}

var sampleDynamicPolicy = DynamicPolicy{
	ID:          "dpol_456",
	Name:        "Rate Limit API",
	Description: "Rate limit API calls",
	Category:    CategoryDynamicCost,
	Tier:        TierOrganization,
	Enabled:     true,
	Config: DynamicPolicyConfig{
		Type:   "rate-limit",
		Rules:  map[string]interface{}{"maxRequestsPerMinute": 100},
		Action: ActionBlock,
	},
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
	Version:   1,
}

var sampleOverride = PolicyOverride{
	PolicyID:  "pol_123",
	Action:    OverrideActionWarn,
	Reason:    "Testing override",
	CreatedAt: time.Now(),
	Active:    true,
}

// TestListStaticPolicies tests listing static policies
func TestListStaticPolicies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/static-policies" {
			t.Errorf("Expected path /api/v1/static-policies, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]StaticPolicy{sampleStaticPolicy})
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	policies, err := client.ListStaticPolicies(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(policies) != 1 {
		t.Errorf("Expected 1 policy, got %d", len(policies))
	}
	if policies[0].ID != "pol_123" {
		t.Errorf("Expected policy ID pol_123, got %s", policies[0].ID)
	}
}

// TestListStaticPoliciesWithFilters tests listing with filters
func TestListStaticPoliciesWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("category") != "security-sqli" {
			t.Errorf("Expected category=security-sqli, got %s", query.Get("category"))
		}
		if query.Get("tier") != "system" {
			t.Errorf("Expected tier=system, got %s", query.Get("tier"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]StaticPolicy{sampleStaticPolicy})
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	enabled := true
	options := &ListStaticPoliciesOptions{
		Category: CategorySecuritySQLI,
		Tier:     TierSystem,
		Enabled:  &enabled,
	}

	policies, err := client.ListStaticPolicies(options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(policies) != 1 {
		t.Errorf("Expected 1 policy, got %d", len(policies))
	}
}

// TestGetStaticPolicy tests getting a specific policy
func TestGetStaticPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/static-policies/pol_123" {
			t.Errorf("Expected path /api/v1/static-policies/pol_123, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleStaticPolicy)
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	policy, err := client.GetStaticPolicy("pol_123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if policy.ID != "pol_123" {
		t.Errorf("Expected policy ID pol_123, got %s", policy.ID)
	}
}

// TestCreateStaticPolicy tests creating a new policy
func TestCreateStaticPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleStaticPolicy)
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	req := &CreateStaticPolicyRequest{
		Name:     "Block SQL Injection",
		Category: CategorySecuritySQLI,
		Pattern:  "(?i)(union\\s+select|drop\\s+table)",
		Severity: 9,
		Enabled:  true,
	}

	policy, err := client.CreateStaticPolicy(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if policy.ID != "pol_123" {
		t.Errorf("Expected policy ID pol_123, got %s", policy.ID)
	}
}

// TestUpdateStaticPolicy tests updating a policy
func TestUpdateStaticPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		updated := sampleStaticPolicy
		updated.Severity = 10
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	severity := 10
	req := &UpdateStaticPolicyRequest{
		Severity: &severity,
	}

	policy, err := client.UpdateStaticPolicy("pol_123", req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if policy.Severity != 10 {
		t.Errorf("Expected severity 10, got %d", policy.Severity)
	}
}

// TestDeleteStaticPolicy tests deleting a policy
func TestDeleteStaticPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	err := client.DeleteStaticPolicy("pol_123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// TestToggleStaticPolicy tests toggling a policy
func TestToggleStaticPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH method, got %s", r.Method)
		}
		toggled := sampleStaticPolicy
		toggled.Enabled = false
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(toggled)
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	policy, err := client.ToggleStaticPolicy("pol_123", false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if policy.Enabled {
		t.Errorf("Expected policy to be disabled")
	}
}

// TestGetEffectiveStaticPolicies tests getting effective policies
func TestGetEffectiveStaticPolicies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/static-policies/effective" {
			t.Errorf("Expected path /api/v1/static-policies/effective, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]StaticPolicy{sampleStaticPolicy})
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	policies, err := client.GetEffectiveStaticPolicies(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(policies) != 1 {
		t.Errorf("Expected 1 policy, got %d", len(policies))
	}
}

// TestTestPattern tests pattern testing
func TestTestPattern(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		result := TestPatternResult{
			Valid: true,
			Results: []TestPatternMatch{
				{Input: "SELECT * FROM users", Matched: true},
				{Input: "Hello world", Matched: false},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	result, err := client.TestPattern("(?i)select", []string{"SELECT * FROM users", "Hello world"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("Expected pattern to be valid")
	}
	if len(result.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result.Results))
	}
}

// TestGetStaticPolicyVersions tests getting version history
func TestGetStaticPolicyVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/static-policies/pol_123/versions" {
			t.Errorf("Expected path /api/v1/static-policies/pol_123/versions, got %s", r.URL.Path)
		}
		versions := []PolicyVersion{
			{Version: 2, ChangeType: "updated", ChangedAt: time.Now()},
			{Version: 1, ChangeType: "created", ChangedAt: time.Now()},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(versions)
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	versions, err := client.GetStaticPolicyVersions("pol_123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(versions))
	}
}

// TestCreatePolicyOverride tests creating an override
func TestCreatePolicyOverride(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/static-policies/pol_123/override" {
			t.Errorf("Expected path /api/v1/static-policies/pol_123/override, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleOverride)
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	req := &CreatePolicyOverrideRequest{
		Action: OverrideActionWarn,
		Reason: "Testing override",
	}

	override, err := client.CreatePolicyOverride("pol_123", req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if override.Action != OverrideActionWarn {
		t.Errorf("Expected action warn, got %s", override.Action)
	}
}

// TestDeletePolicyOverride tests deleting an override
func TestDeletePolicyOverride(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	err := client.DeletePolicyOverride("pol_123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// TestListDynamicPolicies tests listing dynamic policies
func TestListDynamicPolicies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/policies" {
			t.Errorf("Expected path /api/v1/policies, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]DynamicPolicy{sampleDynamicPolicy})
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	policies, err := client.ListDynamicPolicies(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(policies) != 1 {
		t.Errorf("Expected 1 policy, got %d", len(policies))
	}
}

// TestGetDynamicPolicy tests getting a specific dynamic policy
func TestGetDynamicPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/policies/dpol_456" {
			t.Errorf("Expected path /api/v1/policies/dpol_456, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleDynamicPolicy)
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	policy, err := client.GetDynamicPolicy("dpol_456")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if policy.ID != "dpol_456" {
		t.Errorf("Expected policy ID dpol_456, got %s", policy.ID)
	}
}

// TestCreateDynamicPolicy tests creating a dynamic policy
func TestCreateDynamicPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sampleDynamicPolicy)
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	req := &CreateDynamicPolicyRequest{
		Name:     "Rate Limit API",
		Category: CategoryDynamicCost,
		Config: DynamicPolicyConfig{
			Type:   "rate-limit",
			Rules:  map[string]interface{}{"maxRequestsPerMinute": 100},
			Action: ActionBlock,
		},
		Enabled: true,
	}

	policy, err := client.CreateDynamicPolicy(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if policy.ID != "dpol_456" {
		t.Errorf("Expected policy ID dpol_456, got %s", policy.ID)
	}
}

// TestDeleteDynamicPolicy tests deleting a dynamic policy
func TestDeleteDynamicPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	err := client.DeleteDynamicPolicy("dpol_456")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// TestGetEffectiveDynamicPolicies tests getting effective dynamic policies
func TestGetEffectiveDynamicPolicies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/policies/effective" {
			t.Errorf("Expected path /api/v1/policies/effective, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]DynamicPolicy{sampleDynamicPolicy})
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	policies, err := client.GetEffectiveDynamicPolicies(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(policies) != 1 {
		t.Errorf("Expected 1 policy, got %d", len(policies))
	}
}
