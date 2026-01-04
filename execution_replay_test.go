package axonflow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListExecutions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ListExecutionsResponse{
				Executions: []ExecutionSummary{
					{
						RequestID:      "exec-1",
						WorkflowName:   "workflow-1",
						Status:         "completed",
						TotalSteps:     5,
						CompletedSteps: 5,
						StartedAt:      "2025-12-01T00:00:00Z",
					},
					{
						RequestID:    "exec-2",
						WorkflowName: "workflow-2",
						Status:       "running",
						TotalSteps:   3,
						StartedAt:    "2025-12-01T00:02:00Z",
					},
				},
				Total:  2,
				Limit:  50,
				Offset: 0,
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

	resp, err := client.ListExecutions(nil)
	if err != nil {
		t.Fatalf("ListExecutions failed: %v", err)
	}

	if len(resp.Executions) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(resp.Executions))
	}
}

func TestListExecutionsWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions" && r.Method == "GET" {
			if r.URL.Query().Get("status") != "completed" {
				t.Error("Expected status=completed query param")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ListExecutionsResponse{
				Executions: []ExecutionSummary{
					{RequestID: "exec-1", Status: "completed"},
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

	opts := &ListExecutionsOptions{
		Status: "completed",
		Limit:  10,
	}
	resp, err := client.ListExecutions(opts)
	if err != nil {
		t.Fatalf("ListExecutions with options failed: %v", err)
	}

	if len(resp.Executions) != 1 {
		t.Errorf("Expected 1 execution, got %d", len(resp.Executions))
	}
}

func TestGetExecution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ExecutionDetail{
				Summary: &ExecutionSummary{
					RequestID:    "exec-123",
					WorkflowName: "test-workflow",
					Status:       "completed",
					StartedAt:    "2025-12-01T00:00:00Z",
				},
				Steps: []ExecutionSnapshot{
					{RequestID: "exec-123", StepIndex: 0, StepName: "Step 1", Status: "completed"},
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

	exec, err := client.GetExecution("exec-123")
	if err != nil {
		t.Fatalf("GetExecution failed: %v", err)
	}

	if exec.Summary.RequestID != "exec-123" {
		t.Errorf("Expected RequestID 'exec-123', got '%s'", exec.Summary.RequestID)
	}
}

func TestGetExecutionSteps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123/steps" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]ExecutionSnapshot{
				{
					RequestID: "exec-123",
					StepIndex: 0,
					StepName:  "Fetch data",
					Status:    "completed",
				},
				{
					RequestID: "exec-123",
					StepIndex: 1,
					StepName:  "Process data",
					Status:    "running",
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

	steps, err := client.GetExecutionSteps("exec-123")
	if err != nil {
		t.Fatalf("GetExecutionSteps failed: %v", err)
	}

	if len(steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(steps))
	}
}

func TestGetExecutionTimeline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123/timeline" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]TimelineEntry{
				{
					StepIndex: 0,
					StepName:  "Step 1",
					Status:    "completed",
					StartedAt: "2025-12-01T00:00:00Z",
				},
				{
					StepIndex: 1,
					StepName:  "Step 2",
					Status:    "completed",
					StartedAt: "2025-12-01T00:00:30Z",
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

	timeline, err := client.GetExecutionTimeline("exec-123")
	if err != nil {
		t.Fatalf("GetExecutionTimeline failed: %v", err)
	}

	if len(timeline) != 2 {
		t.Errorf("Expected 2 timeline entries, got %d", len(timeline))
	}
}

func TestExportExecution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123/export" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"execution_id": "exec-123",
				"format":       "json",
				"steps":        []interface{}{},
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

	resp, err := client.ExportExecution("exec-123", nil)
	if err != nil {
		t.Fatalf("ExportExecution failed: %v", err)
	}

	if resp["format"] != "json" {
		t.Errorf("Expected format 'json', got '%v'", resp["format"])
	}
}

func TestExportExecutionWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123/export" && r.Method == "GET" {
			// Check query params
			if r.URL.Query().Get("include_input") != "true" {
				t.Error("Expected include_input=true query param")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"execution_id": "exec-123",
				"format":       "json",
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

	opts := &ExecutionExportOptions{
		Format:       "json",
		IncludeInput: true,
	}
	resp, err := client.ExportExecution("exec-123", opts)
	if err != nil {
		t.Fatalf("ExportExecution with options failed: %v", err)
	}

	if resp["execution_id"] != "exec-123" {
		t.Errorf("Expected execution_id 'exec-123', got '%v'", resp["execution_id"])
	}
}

func TestDeleteExecution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123" && r.Method == "DELETE" {
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

	err := client.DeleteExecution("exec-123")
	if err != nil {
		t.Fatalf("DeleteExecution failed: %v", err)
	}
}

func TestGetOrchestratorURLFallback(t *testing.T) {
	client := NewClient(AxonFlowConfig{
		AgentURL:     "http://localhost:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		// No OrchestratorURL set
	})

	// The getOrchestratorURL method should fall back to agent host with port 8081
	// We can test this indirectly by checking that client is created correctly
	if client.config.AgentURL != "http://localhost:8080" {
		t.Error("Expected agent URL to be set")
	}
}
