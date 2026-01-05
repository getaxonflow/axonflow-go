package axonflow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestSearchAuditLogs tests the SearchAuditLogs method
func TestSearchAuditLogs(t *testing.T) {
	t.Run("successful search with all filters", func(t *testing.T) {
		startTime := time.Now().Add(-24 * time.Hour)
		endTime := time.Now()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			if r.URL.Path != "/api/v1/audit/search" {
				t.Errorf("expected path /api/v1/audit/search, got %s", r.URL.Path)
			}
			if r.Method != "POST" {
				t.Errorf("expected method POST, got %s", r.Method)
			}

			// Parse request body
			var reqBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("failed to decode request body: %v", err)
			}

			// Verify filters are passed
			if reqBody["user_email"] != "test@example.com" {
				t.Errorf("expected user_email test@example.com, got %v", reqBody["user_email"])
			}
			if reqBody["client_id"] != "test-client" {
				t.Errorf("expected client_id test-client, got %v", reqBody["client_id"])
			}
			if reqBody["request_type"] != "llm_chat" {
				t.Errorf("expected request_type llm_chat, got %v", reqBody["request_type"])
			}
			if reqBody["limit"].(float64) != 50 {
				t.Errorf("expected limit 50, got %v", reqBody["limit"])
			}

			// Return mock response
			entries := []AuditLogEntry{
				{
					ID:           "audit-1",
					RequestID:    "req-1",
					Timestamp:    time.Now(),
					UserEmail:    "test@example.com",
					ClientID:     "test-client",
					TenantID:     "tenant-1",
					RequestType:  "llm_chat",
					QuerySummary: "Test query",
					Success:      true,
					Blocked:      false,
					RiskScore:    0.1,
					Provider:     "openai",
					Model:        "gpt-4",
					TokensUsed:   150,
					LatencyMs:    250,
				},
				{
					ID:               "audit-2",
					RequestID:        "req-2",
					Timestamp:        time.Now(),
					UserEmail:        "test@example.com",
					ClientID:         "test-client",
					TenantID:         "tenant-1",
					RequestType:      "llm_chat",
					QuerySummary:     "Another query",
					Success:          true,
					Blocked:          true,
					RiskScore:        0.9,
					Provider:         "openai",
					Model:            "gpt-4",
					TokensUsed:       0,
					LatencyMs:        50,
					PolicyViolations: []string{"policy-1"},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entries)
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			AgentURL:        server.URL,
			OrchestratorURL: server.URL,
			ClientID:        "test-client",
			LicenseKey:      "test-key",
		})

		req := &AuditSearchRequest{
			UserEmail:   "test@example.com",
			ClientID:    "test-client",
			StartTime:   &startTime,
			EndTime:     &endTime,
			RequestType: "llm_chat",
			Limit:       50,
		}

		result, err := client.SearchAuditLogs(context.Background(), req)
		if err != nil {
			t.Fatalf("SearchAuditLogs failed: %v", err)
		}

		if len(result.Entries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(result.Entries))
		}
		if result.Entries[0].ID != "audit-1" {
			t.Errorf("expected first entry ID audit-1, got %s", result.Entries[0].ID)
		}
		if result.Entries[1].Blocked != true {
			t.Errorf("expected second entry to be blocked")
		}
		if len(result.Entries[1].PolicyViolations) != 1 {
			t.Errorf("expected 1 policy violation, got %d", len(result.Entries[1].PolicyViolations))
		}
	})

	t.Run("empty search with defaults", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&reqBody)

			// Verify default limit is applied
			if reqBody["limit"].(float64) != 100 {
				t.Errorf("expected default limit 100, got %v", reqBody["limit"])
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]AuditLogEntry{})
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		result, err := client.SearchAuditLogs(context.Background(), nil)
		if err != nil {
			t.Fatalf("SearchAuditLogs failed: %v", err)
		}

		if len(result.Entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(result.Entries))
		}
		if result.Limit != 100 {
			t.Errorf("expected limit 100, got %d", result.Limit)
		}
	})

	t.Run("limit capped at 1000", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&reqBody)

			// Verify limit is capped at 1000
			if reqBody["limit"].(float64) != 1000 {
				t.Errorf("expected limit capped at 1000, got %v", reqBody["limit"])
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]AuditLogEntry{})
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		req := &AuditSearchRequest{Limit: 5000} // Request more than max
		_, err := client.SearchAuditLogs(context.Background(), req)
		if err != nil {
			t.Fatalf("SearchAuditLogs failed: %v", err)
		}
	})

	t.Run("handles 400 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid request"}`))
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		_, err := client.SearchAuditLogs(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for 400 response")
		}

		httpErr, ok := err.(*httpError)
		if !ok {
			t.Fatalf("expected httpError, got %T", err)
		}
		if httpErr.statusCode != 400 {
			t.Errorf("expected status 400, got %d", httpErr.statusCode)
		}
	})

	t.Run("handles 401 unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "unauthorized"}`))
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		_, err := client.SearchAuditLogs(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for 401 response")
		}

		httpErr, ok := err.(*httpError)
		if !ok {
			t.Fatalf("expected httpError, got %T", err)
		}
		if httpErr.statusCode != 401 {
			t.Errorf("expected status 401, got %d", httpErr.statusCode)
		}
	})

	t.Run("handles 500 server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		_, err := client.SearchAuditLogs(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error for 500 response")
		}

		httpErr, ok := err.(*httpError)
		if !ok {
			t.Fatalf("expected httpError, got %T", err)
		}
		if httpErr.statusCode != 500 {
			t.Errorf("expected status 500, got %d", httpErr.statusCode)
		}
	})

	t.Run("handles network error", func(t *testing.T) {
		client := NewClient(AxonFlowConfig{
			OrchestratorURL: "http://localhost:99999", // Invalid port
		})

		_, err := client.SearchAuditLogs(context.Background(), nil)
		if err == nil {
			t.Fatal("expected network error")
		}
	})

	t.Run("handles wrapped response format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := AuditSearchResponse{
				Entries: []AuditLogEntry{
					{ID: "audit-1", RequestID: "req-1"},
				},
				Total:  100,
				Limit:  10,
				Offset: 0,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		result, err := client.SearchAuditLogs(context.Background(), nil)
		if err != nil {
			t.Fatalf("SearchAuditLogs failed: %v", err)
		}

		if result.Total != 100 {
			t.Errorf("expected total 100, got %d", result.Total)
		}
	})

	t.Run("with pagination offset", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqBody map[string]interface{}
			json.NewDecoder(r.Body).Decode(&reqBody)

			if reqBody["offset"].(float64) != 50 {
				t.Errorf("expected offset 50, got %v", reqBody["offset"])
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]AuditLogEntry{})
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		req := &AuditSearchRequest{Offset: 50}
		_, err := client.SearchAuditLogs(context.Background(), req)
		if err != nil {
			t.Fatalf("SearchAuditLogs failed: %v", err)
		}
	})

	t.Run("includes auth headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-License-Key") != "my-license" {
				t.Errorf("expected X-License-Key header")
			}
			if r.Header.Get("X-Client-Secret") != "my-secret" {
				t.Errorf("expected X-Client-Secret header")
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]AuditLogEntry{})
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
			LicenseKey:      "my-license",
			ClientSecret:    "my-secret",
		})

		_, err := client.SearchAuditLogs(context.Background(), nil)
		if err != nil {
			t.Fatalf("SearchAuditLogs failed: %v", err)
		}
	})
}

// TestGetAuditLogsByTenant tests the GetAuditLogsByTenant method
func TestGetAuditLogsByTenant(t *testing.T) {
	t.Run("successful query with defaults", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request
			if r.URL.Path != "/api/v1/audit/tenant/tenant-abc" {
				t.Errorf("expected path /api/v1/audit/tenant/tenant-abc, got %s", r.URL.Path)
			}
			if r.Method != "GET" {
				t.Errorf("expected method GET, got %s", r.Method)
			}

			// Verify default query params
			if r.URL.Query().Get("limit") != "50" {
				t.Errorf("expected default limit 50, got %s", r.URL.Query().Get("limit"))
			}
			if r.URL.Query().Get("offset") != "0" {
				t.Errorf("expected default offset 0, got %s", r.URL.Query().Get("offset"))
			}

			entries := []AuditLogEntry{
				{
					ID:        "audit-1",
					TenantID:  "tenant-abc",
					Success:   true,
					Timestamp: time.Now(),
				},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entries)
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		result, err := client.GetAuditLogsByTenant(context.Background(), "tenant-abc", nil)
		if err != nil {
			t.Fatalf("GetAuditLogsByTenant failed: %v", err)
		}

		if len(result.Entries) != 1 {
			t.Errorf("expected 1 entry, got %d", len(result.Entries))
		}
		if result.Limit != 50 {
			t.Errorf("expected limit 50, got %d", result.Limit)
		}
	})

	t.Run("with custom options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("limit") != "100" {
				t.Errorf("expected limit 100, got %s", r.URL.Query().Get("limit"))
			}
			if r.URL.Query().Get("offset") != "25" {
				t.Errorf("expected offset 25, got %s", r.URL.Query().Get("offset"))
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]AuditLogEntry{})
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		opts := &AuditQueryOptions{
			Limit:  100,
			Offset: 25,
		}

		_, err := client.GetAuditLogsByTenant(context.Background(), "tenant-abc", opts)
		if err != nil {
			t.Fatalf("GetAuditLogsByTenant failed: %v", err)
		}
	})

	t.Run("empty tenant ID returns error", func(t *testing.T) {
		client := NewClient(AxonFlowConfig{
			OrchestratorURL: "http://localhost:8081",
		})

		_, err := client.GetAuditLogsByTenant(context.Background(), "", nil)
		if err == nil {
			t.Fatal("expected error for empty tenant ID")
		}
		if err.Error() != "tenantID is required" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("limit capped at 1000", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("limit") != "1000" {
				t.Errorf("expected limit capped at 1000, got %s", r.URL.Query().Get("limit"))
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]AuditLogEntry{})
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		opts := &AuditQueryOptions{Limit: 5000}
		_, err := client.GetAuditLogsByTenant(context.Background(), "tenant-abc", opts)
		if err != nil {
			t.Fatalf("GetAuditLogsByTenant failed: %v", err)
		}
	})

	t.Run("handles 400 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid tenant ID"}`))
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		_, err := client.GetAuditLogsByTenant(context.Background(), "invalid!", nil)
		if err == nil {
			t.Fatal("expected error for 400 response")
		}

		httpErr, ok := err.(*httpError)
		if !ok {
			t.Fatalf("expected httpError, got %T", err)
		}
		if httpErr.statusCode != 400 {
			t.Errorf("expected status 400, got %d", httpErr.statusCode)
		}
	})

	t.Run("handles 403 forbidden", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "not authorized for this tenant"}`))
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		_, err := client.GetAuditLogsByTenant(context.Background(), "other-tenant", nil)
		if err == nil {
			t.Fatal("expected error for 403 response")
		}

		httpErr, ok := err.(*httpError)
		if !ok {
			t.Fatalf("expected httpError, got %T", err)
		}
		if httpErr.statusCode != 403 {
			t.Errorf("expected status 403, got %d", httpErr.statusCode)
		}
	})

	t.Run("handles 404 not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "tenant not found"}`))
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		_, err := client.GetAuditLogsByTenant(context.Background(), "nonexistent", nil)
		if err == nil {
			t.Fatal("expected error for 404 response")
		}

		httpErr, ok := err.(*httpError)
		if !ok {
			t.Fatalf("expected httpError, got %T", err)
		}
		if httpErr.statusCode != 404 {
			t.Errorf("expected status 404, got %d", httpErr.statusCode)
		}
	})

	t.Run("handles network error", func(t *testing.T) {
		client := NewClient(AxonFlowConfig{
			OrchestratorURL: "http://localhost:99999",
		})

		_, err := client.GetAuditLogsByTenant(context.Background(), "tenant-abc", nil)
		if err == nil {
			t.Fatal("expected network error")
		}
	})

	t.Run("handles wrapped response format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := AuditSearchResponse{
				Entries: []AuditLogEntry{
					{ID: "audit-1", TenantID: "tenant-abc"},
				},
				Total:  50,
				Limit:  50,
				Offset: 0,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
		})

		result, err := client.GetAuditLogsByTenant(context.Background(), "tenant-abc", nil)
		if err != nil {
			t.Fatalf("GetAuditLogsByTenant failed: %v", err)
		}

		if result.Total != 50 {
			t.Errorf("expected total 50, got %d", result.Total)
		}
	})

	t.Run("includes auth headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-License-Key") != "my-license" {
				t.Errorf("expected X-License-Key header")
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]AuditLogEntry{})
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
			LicenseKey:      "my-license",
		})

		_, err := client.GetAuditLogsByTenant(context.Background(), "tenant-abc", nil)
		if err != nil {
			t.Fatalf("GetAuditLogsByTenant failed: %v", err)
		}
	})

	t.Run("debug mode logging", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]AuditLogEntry{})
		}))
		defer server.Close()

		client := NewClient(AxonFlowConfig{
			OrchestratorURL: server.URL,
			Debug:           true,
		})

		_, err := client.GetAuditLogsByTenant(context.Background(), "tenant-abc", nil)
		if err != nil {
			t.Fatalf("GetAuditLogsByTenant failed: %v", err)
		}
	})
}

// TestAuditLogEntryFields tests that AuditLogEntry correctly unmarshals all fields
func TestAuditLogEntryFields(t *testing.T) {
	jsonData := `{
		"id": "audit-123",
		"request_id": "req-456",
		"timestamp": "2026-01-05T10:30:00Z",
		"user_email": "user@example.com",
		"client_id": "client-1",
		"tenant_id": "tenant-1",
		"request_type": "llm_chat",
		"query_summary": "What is AI?",
		"success": true,
		"blocked": false,
		"risk_score": 0.25,
		"provider": "openai",
		"model": "gpt-4",
		"tokens_used": 500,
		"latency_ms": 1200,
		"policy_violations": ["pol-1", "pol-2"],
		"metadata": {"key": "value"}
	}`

	var entry AuditLogEntry
	if err := json.Unmarshal([]byte(jsonData), &entry); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if entry.ID != "audit-123" {
		t.Errorf("expected ID audit-123, got %s", entry.ID)
	}
	if entry.RequestID != "req-456" {
		t.Errorf("expected RequestID req-456, got %s", entry.RequestID)
	}
	if entry.UserEmail != "user@example.com" {
		t.Errorf("expected UserEmail user@example.com, got %s", entry.UserEmail)
	}
	if entry.RiskScore != 0.25 {
		t.Errorf("expected RiskScore 0.25, got %f", entry.RiskScore)
	}
	if entry.TokensUsed != 500 {
		t.Errorf("expected TokensUsed 500, got %d", entry.TokensUsed)
	}
	if len(entry.PolicyViolations) != 2 {
		t.Errorf("expected 2 policy violations, got %d", len(entry.PolicyViolations))
	}
	if entry.Metadata["key"] != "value" {
		t.Errorf("expected metadata key=value")
	}
}

// TestAuditSearchRequestSerialization tests request serialization
func TestAuditSearchRequestSerialization(t *testing.T) {
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 1, 5, 23, 59, 59, 0, time.UTC)

	req := AuditSearchRequest{
		UserEmail:   "test@example.com",
		ClientID:    "client-1",
		StartTime:   &startTime,
		EndTime:     &endTime,
		RequestType: "llm_chat",
		Limit:       50,
		Offset:      10,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["user_email"] != "test@example.com" {
		t.Errorf("expected user_email test@example.com")
	}
	if parsed["limit"].(float64) != 50 {
		t.Errorf("expected limit 50")
	}
}

// TestContextCancellation tests that context cancellation is respected
func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]AuditLogEntry{})
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		OrchestratorURL: server.URL,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.SearchAuditLogs(ctx, nil)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
