package axonflow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateBudget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/budgets" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Budget{
				ID:       "budget-123",
				Name:     "Test Budget",
				Scope:    "organization",
				LimitUSD: 100.0,
				Period:   "monthly",
				OnExceed: "warn",
				Enabled:  true,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	budget, err := client.CreateBudget(context.Background(), CreateBudgetRequest{
		Name:     "Test Budget",
		Scope:    "organization",
		LimitUSD: 100.0,
		Period:   "monthly",
		OnExceed: "warn",
	})
	if err != nil {
		t.Fatalf("CreateBudget failed: %v", err)
	}

	if budget.ID != "budget-123" {
		t.Errorf("Expected ID 'budget-123', got '%s'", budget.ID)
	}
}

func TestGetBudget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/budgets/budget-123" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Budget{
				ID:       "budget-123",
				Name:     "Test Budget",
				Scope:    "team",
				LimitUSD: 50.0,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	budget, err := client.GetBudget(context.Background(), "budget-123")
	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}

	if budget.ID != "budget-123" {
		t.Errorf("Expected ID 'budget-123', got '%s'", budget.ID)
	}
}

func TestListBudgets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/budgets" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(BudgetsResponse{
				Budgets: []Budget{
					{ID: "budget-1", Name: "Budget 1"},
					{ID: "budget-2", Name: "Budget 2"},
				},
				Total: 2,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	resp, err := client.ListBudgets(context.Background(), ListBudgetsOptions{})
	if err != nil {
		t.Fatalf("ListBudgets failed: %v", err)
	}

	if len(resp.Budgets) != 2 {
		t.Errorf("Expected 2 budgets, got %d", len(resp.Budgets))
	}
}

func TestListBudgetsWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/budgets" && r.Method == "GET" {
			// Verify query params
			if r.URL.Query().Get("scope") != "team" {
				t.Error("Expected scope=team query param")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(BudgetsResponse{
				Budgets: []Budget{{ID: "budget-1"}},
				Total:   1,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	opts := ListBudgetsOptions{
		Scope: "team",
		Limit: 10,
	}
	resp, err := client.ListBudgets(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListBudgets with options failed: %v", err)
	}

	if len(resp.Budgets) != 1 {
		t.Errorf("Expected 1 budget, got %d", len(resp.Budgets))
	}
}

func TestUpdateBudget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/budgets/budget-123" && r.Method == "PUT" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Budget{
				ID:       "budget-123",
				Name:     "Updated Budget",
				LimitUSD: 200.0,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	budget, err := client.UpdateBudget(context.Background(), &Budget{
		ID:       "budget-123",
		Name:     "Updated Budget",
		LimitUSD: 200.0,
	})
	if err != nil {
		t.Fatalf("UpdateBudget failed: %v", err)
	}

	if budget.Name != "Updated Budget" {
		t.Errorf("Expected name 'Updated Budget', got '%s'", budget.Name)
	}
}

func TestDeleteBudget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/budgets/budget-123" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	err := client.DeleteBudget(context.Background(), "budget-123")
	if err != nil {
		t.Fatalf("DeleteBudget failed: %v", err)
	}
}

func TestGetBudgetStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/budgets/budget-123/status" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(BudgetStatus{
				Budget: Budget{
					ID:       "budget-123",
					LimitUSD: 100.0,
				},
				UsedUSD:      45.50,
				RemainingUSD: 54.50,
				Percentage:   45.5,
				IsExceeded:   false,
				IsBlocked:    false,
				PeriodStart:  "2025-12-01T00:00:00Z",
				PeriodEnd:    "2025-12-31T23:59:59Z",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	status, err := client.GetBudgetStatus(context.Background(), "budget-123")
	if err != nil {
		t.Fatalf("GetBudgetStatus failed: %v", err)
	}

	if status.UsedUSD != 45.50 {
		t.Errorf("Expected UsedUSD 45.50, got %f", status.UsedUSD)
	}
}

func TestGetBudgetAlerts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/budgets/budget-123/alerts" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(BudgetAlertsResponse{
				Alerts: []BudgetAlert{
					{
						ID:                "alert-1",
						BudgetID:          "budget-123",
						AlertType:         "threshold",
						Threshold:         50,
						PercentageReached: 51.2,
						AmountUSD:         51.20,
						Message:           "Threshold reached",
					},
				},
				Count: 1,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	resp, err := client.GetBudgetAlerts(context.Background(), "budget-123", 10)
	if err != nil {
		t.Fatalf("GetBudgetAlerts failed: %v", err)
	}

	if len(resp.Alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(resp.Alerts))
	}
}

func TestCheckBudget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/budgets/check" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(BudgetDecision{
				Allowed: true,
				Action:  "allow",
				Message: "Within budget",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	decision, err := client.CheckBudget(context.Background(), CheckBudgetRequest{
		OrgID:  "org-1",
		TeamID: "team-1",
	})
	if err != nil {
		t.Fatalf("CheckBudget failed: %v", err)
	}

	if !decision.Allowed {
		t.Error("Expected decision to be allowed")
	}
}

func TestGetUsageSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/usage" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(UsageSummary{
				TotalCostUSD:          150.75,
				TotalRequests:         5000,
				TotalTokensIn:         1000000,
				TotalTokensOut:        500000,
				AverageCostPerRequest: 0.03,
				Period:                "monthly",
				PeriodStart:           "2025-12-01T00:00:00Z",
				PeriodEnd:             "2025-12-31T23:59:59Z",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	summary, err := client.GetUsageSummary(context.Background(), UsageQueryOptions{})
	if err != nil {
		t.Fatalf("GetUsageSummary failed: %v", err)
	}

	if summary.TotalCostUSD != 150.75 {
		t.Errorf("Expected TotalCostUSD 150.75, got %f", summary.TotalCostUSD)
	}
}

func TestGetUsageBreakdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/usage/breakdown" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(UsageBreakdown{
				GroupBy:      "provider",
				TotalCostUSD: 150.75,
				Items: []UsageBreakdownItem{
					{GroupValue: "openai", CostUSD: 100.0, Percentage: 66.7},
					{GroupValue: "anthropic", CostUSD: 50.75, Percentage: 33.3},
				},
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	breakdown, err := client.GetUsageBreakdown(context.Background(), "provider", UsageQueryOptions{})
	if err != nil {
		t.Fatalf("GetUsageBreakdown failed: %v", err)
	}

	if len(breakdown.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(breakdown.Items))
	}
}

func TestListUsageRecords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/usage/records" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(UsageRecordsResponse{
				Records: []UsageRecord{
					{
						ID:        "record-1",
						Provider:  "openai",
						Model:     "gpt-4",
						TokensIn:  100,
						TokensOut: 50,
						CostUSD:   0.0045,
					},
				},
				Total: 1,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	resp, err := client.ListUsageRecords(context.Background(), UsageQueryOptions{})
	if err != nil {
		t.Fatalf("ListUsageRecords failed: %v", err)
	}

	if len(resp.Records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(resp.Records))
	}
}

func TestGetPricing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/pricing" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PricingInfo{
				Provider: "openai",
				Model:    "gpt-4",
				Pricing: ModelPricing{
					InputPer1K:  0.03,
					OutputPer1K: 0.06,
				},
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	resp, err := client.GetPricing(context.Background(), "openai", "gpt-4")
	if err != nil {
		t.Fatalf("GetPricing failed: %v", err)
	}

	if resp.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", resp.Provider)
	}
}

func TestLastIndex(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		sep      string
		expected int
	}{
		{"colon in URL", "http://localhost:8080", ":", 16},
		{"no match", "http://localhost", "x", -1},
		{"empty separator", "hello", "", 5},
		{"empty string", "", ":", -1},
		{"multiple matches", "a:b:c:d", ":", 5},
		{"single character match", "hello", "o", 4},
		{"separator at start", ":hello", ":", 0},
		{"separator at end", "hello:", ":", 5},
		{"long separator", "hello world world", "world", 12},
		{"slash in URL", "http://localhost:8080/path/to/resource", "/", 29},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lastIndex(tt.s, tt.sep)
			if result != tt.expected {
				t.Errorf("lastIndex(%q, %q) = %d, want %d", tt.s, tt.sep, result, tt.expected)
			}
		})
	}
}

// Note: Tests for OrchestratorURL fallback were removed in v2.0.0 (ADR-026 Single Entry Point).
// All routes now go through the single Endpoint field.

func TestCostRequestUsesEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/budgets/test-budget" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(Budget{
				ID:       "test-budget",
				Name:     "Test Budget",
				LimitUSD: 100.0,
			})
		}
	}))
	defer server.Close()

	// All routes go through single endpoint
	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	budget, err := client.GetBudget(context.Background(), "test-budget")
	if err != nil {
		t.Fatalf("GetBudget failed: %v", err)
	}

	if budget.ID != "test-budget" {
		t.Errorf("Expected budget ID 'test-budget', got '%s'", budget.ID)
	}
}

func TestCostRequestWithEmptyEndpoint(t *testing.T) {
	client := NewClient(AxonFlowConfig{
		Endpoint:     "",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	// This exercises the empty endpoint path
	_, err := client.GetBudget(context.Background(), "test-budget")
	if err == nil {
		t.Error("Expected error when Endpoint is empty")
	}
}

func TestListBudgetsWithAllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/budgets" && r.Method == "GET" {
			// Verify all query params
			if r.URL.Query().Get("scope") != "organization" {
				t.Error("Expected scope=organization query param")
			}
			if r.URL.Query().Get("limit") != "25" {
				t.Error("Expected limit=25 query param")
			}
			if r.URL.Query().Get("offset") != "10" {
				t.Error("Expected offset=10 query param")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(BudgetsResponse{
				Budgets: []Budget{{ID: "budget-1"}},
				Total:   1,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	opts := ListBudgetsOptions{
		Scope:  "organization",
		Limit:  25,
		Offset: 10,
	}
	resp, err := client.ListBudgets(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListBudgets with all options failed: %v", err)
	}

	if len(resp.Budgets) != 1 {
		t.Errorf("Expected 1 budget, got %d", len(resp.Budgets))
	}
}

func TestGetBudgetAlertsWithNoLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/budgets/budget-123/alerts" && r.Method == "GET" {
			// Should not have limit param when limit is 0
			if r.URL.Query().Get("limit") != "" {
				t.Error("Expected no limit query param when limit is 0")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(BudgetAlertsResponse{
				Alerts: nil, // Test null handling
				Count:  0,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	resp, err := client.GetBudgetAlerts(context.Background(), "budget-123", 0)
	if err != nil {
		t.Fatalf("GetBudgetAlerts failed: %v", err)
	}

	// Should handle null alerts
	if resp.Alerts == nil {
		t.Error("Expected Alerts to be initialized to empty slice, got nil")
	}
}

func TestGetUsageSummaryWithPeriod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/usage" && r.Method == "GET" {
			// Verify period param
			if r.URL.Query().Get("period") != "weekly" {
				t.Error("Expected period=weekly query param")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(UsageSummary{
				TotalCostUSD: 50.0,
				Period:       "weekly",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	summary, err := client.GetUsageSummary(context.Background(), UsageQueryOptions{
		Period: "weekly",
	})
	if err != nil {
		t.Fatalf("GetUsageSummary with period failed: %v", err)
	}

	if summary.Period != "weekly" {
		t.Errorf("Expected period 'weekly', got '%s'", summary.Period)
	}
}

func TestGetUsageBreakdownWithPeriod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/usage/breakdown" && r.Method == "GET" {
			// Verify query params
			if r.URL.Query().Get("group_by") != "model" {
				t.Error("Expected group_by=model query param")
			}
			if r.URL.Query().Get("period") != "daily" {
				t.Error("Expected period=daily query param")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(UsageBreakdown{
				GroupBy: "model",
				Items:   nil, // Test null handling
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	breakdown, err := client.GetUsageBreakdown(context.Background(), "model", UsageQueryOptions{
		Period: "daily",
	})
	if err != nil {
		t.Fatalf("GetUsageBreakdown with period failed: %v", err)
	}

	// Should handle null items
	if breakdown.Items == nil {
		t.Error("Expected Items to be initialized to empty slice, got nil")
	}
}

func TestListUsageRecordsWithAllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/usage/records" && r.Method == "GET" {
			// Verify all query params
			if r.URL.Query().Get("limit") != "50" {
				t.Error("Expected limit=50 query param")
			}
			if r.URL.Query().Get("offset") != "25" {
				t.Error("Expected offset=25 query param")
			}
			if r.URL.Query().Get("provider") != "anthropic" {
				t.Error("Expected provider=anthropic query param")
			}
			if r.URL.Query().Get("model") != "claude-3" {
				t.Error("Expected model=claude-3 query param")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(UsageRecordsResponse{
				Records: nil, // Test null handling
				Total:   0,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	resp, err := client.ListUsageRecords(context.Background(), UsageQueryOptions{
		Limit:    50,
		Offset:   25,
		Provider: "anthropic",
		Model:    "claude-3",
	})
	if err != nil {
		t.Fatalf("ListUsageRecords with all options failed: %v", err)
	}

	// Should handle null records
	if resp.Records == nil {
		t.Error("Expected Records to be initialized to empty slice, got nil")
	}
}

func TestGetPricingWithNoParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/pricing" && r.Method == "GET" {
			// Should not have provider or model params
			if r.URL.Query().Get("provider") != "" {
				t.Error("Expected no provider query param")
			}
			if r.URL.Query().Get("model") != "" {
				t.Error("Expected no model query param")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PricingInfo{
				Provider: "default",
				Model:    "default",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:        server.URL,
		ClientID:        "test-client",
		ClientSecret:    "test-secret",
	})

	resp, err := client.GetPricing(context.Background(), "", "")
	if err != nil {
		t.Fatalf("GetPricing with no params failed: %v", err)
	}

	if resp.Provider != "default" {
		t.Errorf("Expected provider 'default', got '%s'", resp.Provider)
	}
}
