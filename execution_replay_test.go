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
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
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
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
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
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
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
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
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
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
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
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
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
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
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
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	err := client.DeleteExecution("exec-123")
	if err != nil {
		t.Fatalf("DeleteExecution failed: %v", err)
	}
}

// Note: Tests for getOrchestratorURL fallback were removed in v2.0.0 (ADR-026 Single Entry Point).
// All routes now go through the single Endpoint field.

func TestEndpointUsedForExecutionReplay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ListExecutionsResponse{
				Executions: []ExecutionSummary{},
				Total:      0,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL, // All routes go through single endpoint
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	resp, err := client.ListExecutions(nil)
	if err != nil {
		t.Fatalf("ListExecutions failed: %v", err)
	}

	if resp.Total != 0 {
		t.Errorf("Expected total 0, got %d", resp.Total)
	}
}

func TestListExecutionsWithAllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions" && r.Method == "GET" {
			// Verify all query params
			if r.URL.Query().Get("limit") != "25" {
				t.Error("Expected limit=25 query param")
			}
			if r.URL.Query().Get("offset") != "10" {
				t.Error("Expected offset=10 query param")
			}
			if r.URL.Query().Get("status") != "running" {
				t.Error("Expected status=running query param")
			}
			if r.URL.Query().Get("workflow_id") != "workflow-1" {
				t.Error("Expected workflow_id=workflow-1 query param")
			}
			if r.URL.Query().Get("start_time") != "2025-01-01T00:00:00Z" {
				t.Error("Expected start_time query param")
			}
			if r.URL.Query().Get("end_time") != "2025-12-31T23:59:59Z" {
				t.Error("Expected end_time query param")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ListExecutionsResponse{
				Executions: []ExecutionSummary{{RequestID: "exec-1", Status: "running"}},
				Total:      1,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	opts := &ListExecutionsOptions{
		Limit:      25,
		Offset:     10,
		Status:     "running",
		WorkflowID: "workflow-1",
		StartTime:  "2025-01-01T00:00:00Z",
		EndTime:    "2025-12-31T23:59:59Z",
	}
	resp, err := client.ListExecutions(opts)
	if err != nil {
		t.Fatalf("ListExecutions with all options failed: %v", err)
	}

	if len(resp.Executions) != 1 {
		t.Errorf("Expected 1 execution, got %d", len(resp.Executions))
	}
}

func TestListExecutionsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.ListExecutions(nil)
	if err == nil {
		t.Error("Expected error from server")
	}
}

func TestGetExecutionNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.GetExecution("exec-nonexistent")
	if err == nil {
		t.Error("Expected not found error")
	}
}

func TestGetExecutionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.GetExecution("exec-123")
	if err == nil {
		t.Error("Expected error from server")
	}
}

func TestGetExecutionStepsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.GetExecutionSteps("exec-nonexistent")
	if err == nil {
		t.Error("Expected not found error")
	}
}

func TestGetExecutionStepsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.GetExecutionSteps("exec-123")
	if err == nil {
		t.Error("Expected error from server")
	}
}

func TestGetExecutionTimelineNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.GetExecutionTimeline("exec-nonexistent")
	if err == nil {
		t.Error("Expected not found error")
	}
}

func TestGetExecutionTimelineError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.GetExecutionTimeline("exec-123")
	if err == nil {
		t.Error("Expected error from server")
	}
}

func TestExportExecutionNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.ExportExecution("exec-nonexistent", nil)
	if err == nil {
		t.Error("Expected not found error")
	}
}

func TestExportExecutionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	_, err := client.ExportExecution("exec-123", nil)
	if err == nil {
		t.Error("Expected error from server")
	}
}

func TestExportExecutionWithAllOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123/export" && r.Method == "GET" {
			// Check all query params
			if r.URL.Query().Get("format") != "json" {
				t.Error("Expected format=json query param")
			}
			if r.URL.Query().Get("include_input") != "true" {
				t.Error("Expected include_input=true query param")
			}
			if r.URL.Query().Get("include_output") != "true" {
				t.Error("Expected include_output=true query param")
			}
			if r.URL.Query().Get("include_policies") != "true" {
				t.Error("Expected include_policies=true query param")
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
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	opts := &ExecutionExportOptions{
		Format:          "json",
		IncludeInput:    true,
		IncludeOutput:   true,
		IncludePolicies: true,
	}
	resp, err := client.ExportExecution("exec-123", opts)
	if err != nil {
		t.Fatalf("ExportExecution with all options failed: %v", err)
	}

	if resp["execution_id"] != "exec-123" {
		t.Errorf("Expected execution_id 'exec-123', got '%v'", resp["execution_id"])
	}
}

func TestDeleteExecutionNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	err := client.DeleteExecution("exec-nonexistent")
	if err == nil {
		t.Error("Expected not found error")
	}
}

func TestDeleteExecutionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	err := client.DeleteExecution("exec-123")
	if err == nil {
		t.Error("Expected error from server")
	}
}

func TestDeleteExecutionOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123" && r.Method == "DELETE" {
			// Some APIs return 200 OK instead of 204 No Content
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	err := client.DeleteExecution("exec-123")
	if err != nil {
		t.Fatalf("DeleteExecution failed: %v", err)
	}
}

func TestListExecutionsWithDebugMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ListExecutionsResponse{
				Executions: []ExecutionSummary{{RequestID: "exec-1"}},
				Total:      1,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Debug:        true, // Enable debug mode
	})

	resp, err := client.ListExecutions(nil)
	if err != nil {
		t.Fatalf("ListExecutions failed: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("Expected total 1, got %d", resp.Total)
	}
}

func TestGetExecutionWithDebugMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ExecutionDetail{
				Summary: &ExecutionSummary{
					RequestID: "exec-123",
					Status:    "completed",
				},
				Steps: []ExecutionSnapshot{},
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Debug:        true,
	})

	exec, err := client.GetExecution("exec-123")
	if err != nil {
		t.Fatalf("GetExecution failed: %v", err)
	}

	if exec.Summary.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", exec.Summary.Status)
	}
}

func TestGetExecutionStepsWithDebugMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123/steps" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]ExecutionSnapshot{
				{RequestID: "exec-123", StepIndex: 0, Status: "completed"},
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Debug:        true,
	})

	steps, err := client.GetExecutionSteps("exec-123")
	if err != nil {
		t.Fatalf("GetExecutionSteps failed: %v", err)
	}

	if len(steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(steps))
	}
}

func TestGetExecutionTimelineWithDebugMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123/timeline" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]TimelineEntry{
				{StepIndex: 0, StepName: "Step 1", Status: "completed"},
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Debug:        true,
	})

	timeline, err := client.GetExecutionTimeline("exec-123")
	if err != nil {
		t.Fatalf("GetExecutionTimeline failed: %v", err)
	}

	if len(timeline) != 1 {
		t.Errorf("Expected 1 timeline entry, got %d", len(timeline))
	}
}

func TestExportExecutionWithDebugMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123/export" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"execution_id": "exec-123",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Debug:        true,
	})

	resp, err := client.ExportExecution("exec-123", nil)
	if err != nil {
		t.Fatalf("ExportExecution failed: %v", err)
	}

	if resp["execution_id"] != "exec-123" {
		t.Errorf("Expected execution_id 'exec-123', got '%v'", resp["execution_id"])
	}
}

func TestDeleteExecutionWithDebugMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/executions/exec-123" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		Endpoint:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Debug:        true,
	})

	err := client.DeleteExecution("exec-123")
	if err != nil {
		t.Fatalf("DeleteExecution failed: %v", err)
	}
}
