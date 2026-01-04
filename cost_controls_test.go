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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
		AgentURL:        server.URL,
		OrchestratorURL: server.URL,
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
