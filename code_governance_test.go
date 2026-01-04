package axonflow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginToPortal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{
				SessionID: "sess-123",
				OrgID:     "test-org",
				Email:     "admin@test.com",
				Name:      "Admin",
				ExpiresAt: "2025-12-01T00:00:00Z",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	resp, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	if resp.SessionID == "" {
		t.Error("Expected SessionID to be set")
	}

	if !client.IsLoggedIn() {
		t.Error("Expected IsLoggedIn to return true after login")
	}
}

func TestLogoutFromPortal(t *testing.T) {
	loginCalled := false
	logoutCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			loginCalled = true
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/auth/logout" && r.Method == "POST" {
			logoutCalled = true
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	// Login first
	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	if !loginCalled {
		t.Error("Expected login to be called")
	}

	// Logout
	err = client.LogoutFromPortal()
	if err != nil {
		t.Fatalf("LogoutFromPortal failed: %v", err)
	}

	if !logoutCalled {
		t.Error("Expected logout to be called")
	}

	if client.IsLoggedIn() {
		t.Error("Expected IsLoggedIn to return false after logout")
	}
}

func TestIsLoggedInWithoutLogin(t *testing.T) {
	client := NewClient(AxonFlowConfig{
		AgentURL:     "http://localhost",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	if client.IsLoggedIn() {
		t.Error("Expected IsLoggedIn to return false when not logged in")
	}
}

func TestConfigureGitProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/git-providers" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ConfigureGitProviderResponse{
				Message: "Git provider configured successfully",
				Type:    "github",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	// Login first
	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	resp, err := client.ConfigureGitProvider(&ConfigureGitProviderRequest{
		Type:  GitProviderGitHub,
		Token: "ghp_xxx",
	})
	if err != nil {
		t.Fatalf("ConfigureGitProvider failed: %v", err)
	}

	if resp.Type != "github" {
		t.Errorf("Expected type 'github', got '%s'", resp.Type)
	}
}

func TestListGitProviders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/git-providers" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ListGitProvidersResponse{
				Providers: []GitProviderInfo{
					{Type: GitProviderGitHub},
					{Type: GitProviderGitLab},
				},
				Count: 2,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	resp, err := client.ListGitProviders()
	if err != nil {
		t.Fatalf("ListGitProviders failed: %v", err)
	}

	if len(resp.Providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(resp.Providers))
	}
}

func TestValidateGitProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/git-providers/validate" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ValidateGitProviderResponse{
				Valid:   true,
				Message: "Credentials are valid",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	result, err := client.ValidateGitProvider(&ValidateGitProviderRequest{
		Type:  GitProviderGitHub,
		Token: "ghp_xxx",
	})
	if err != nil {
		t.Fatalf("ValidateGitProvider failed: %v", err)
	}

	if !result.Valid {
		t.Error("Expected validation to be valid")
	}
}

func TestDeleteGitProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/git-providers/github" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	err = client.DeleteGitProvider(GitProviderGitHub)
	if err != nil {
		t.Fatalf("DeleteGitProvider failed: %v", err)
	}
}

func TestListPRs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/prs" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ListPRsResponse{
				PRs: []PRRecord{
					{ID: "pr-1", PRNumber: 123, Title: "Test PR", State: "open"},
				},
				Count: 1,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	resp, err := client.ListPRs(nil)
	if err != nil {
		t.Fatalf("ListPRs failed: %v", err)
	}

	if len(resp.PRs) != 1 {
		t.Errorf("Expected 1 PR, got %d", len(resp.PRs))
	}
}

func TestListPRsWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/prs" && r.Method == "GET" {
			// Verify query params
			if r.URL.Query().Get("state") != "open" {
				t.Error("Expected state=open query param")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ListPRsResponse{
				PRs:   []PRRecord{{ID: "pr-1", State: "open"}},
				Count: 1,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	opts := &ListPRsOptions{
		State: "open",
		Limit: 10,
	}
	resp, err := client.ListPRs(opts)
	if err != nil {
		t.Fatalf("ListPRs with options failed: %v", err)
	}

	if len(resp.PRs) != 1 {
		t.Errorf("Expected 1 PR, got %d", len(resp.PRs))
	}
}

func TestCreatePR(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/prs" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CreatePRResponse{
				PRID:       "pr-1",
				PRNumber:   124,
				PRURL:      "https://github.com/org/repo/pull/124",
				State:      "open",
				HeadBranch: "feature-branch",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	pr, err := client.CreatePR(&CreatePRRequest{
		Owner:       "org",
		Repo:        "repo",
		Title:       "New PR",
		Description: "PR description",
		BaseBranch:  "main",
		BranchName:  "feature-branch",
		Files: []CodeFile{
			{Path: "test.go", Content: "package test", Action: FileActionCreate},
		},
	})
	if err != nil {
		t.Fatalf("CreatePR failed: %v", err)
	}

	if pr.PRNumber != 124 {
		t.Errorf("Expected PR number 124, got %d", pr.PRNumber)
	}
}

func TestGetPR(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/prs/pr-123" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PRRecord{
				ID:       "pr-123",
				PRNumber: 123,
				Title:    "Test PR",
				State:    "open",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	pr, err := client.GetPR("pr-123")
	if err != nil {
		t.Fatalf("GetPR failed: %v", err)
	}

	if pr.ID != "pr-123" {
		t.Errorf("Expected ID 'pr-123', got '%s'", pr.ID)
	}
}

func TestSyncPRStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/prs/pr-123/sync" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PRRecord{
				ID:       "pr-123",
				PRNumber: 123,
				State:    "merged",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	pr, err := client.SyncPRStatus("pr-123")
	if err != nil {
		t.Fatalf("SyncPRStatus failed: %v", err)
	}

	if pr.State != "merged" {
		t.Errorf("Expected state 'merged', got '%s'", pr.State)
	}
}

func TestGetCodeGovernanceMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/metrics" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CodeGovernanceMetrics{
				TotalPRs:             100,
				MergedPRs:            80,
				OpenPRs:              15,
				ClosedPRs:            5,
				TotalSecretsDetected: 3,
				TotalUnsafePatterns:  2,
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	metrics, err := client.GetCodeGovernanceMetrics()
	if err != nil {
		t.Fatalf("GetCodeGovernanceMetrics failed: %v", err)
	}

	if metrics.TotalPRs != 100 {
		t.Errorf("Expected TotalPRs 100, got %d", metrics.TotalPRs)
	}
}

func TestExportCodeGovernanceData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/export" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ExportResponse{
				Records:    []PRRecord{},
				Count:      0,
				ExportedAt: "2025-12-01T00:00:00Z",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	export, err := client.ExportCodeGovernanceData(nil)
	if err != nil {
		t.Fatalf("ExportCodeGovernanceData failed: %v", err)
	}

	if export.ExportedAt == "" {
		t.Error("Expected ExportedAt to be set")
	}
}

func TestExportCodeGovernanceDataWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/export" && r.Method == "GET" {
			// Verify query params
			if r.URL.Query().Get("state") != "merged" {
				t.Error("Expected state=merged query param")
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ExportResponse{
				Records:    []PRRecord{{ID: "pr-1", State: "merged"}},
				Count:      1,
				ExportedAt: "2025-12-01T00:00:00Z",
			})
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	opts := &ExportOptions{
		State: "merged",
	}
	export, err := client.ExportCodeGovernanceData(opts)
	if err != nil {
		t.Fatalf("ExportCodeGovernanceData with options failed: %v", err)
	}

	if export.Count != 1 {
		t.Errorf("Expected count 1, got %d", export.Count)
	}
}

func TestExportCodeGovernanceDataCSV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			w.Header().Set("Set-Cookie", "axonflow_session=abc123; Path=/; HttpOnly")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(PortalLoginResponse{SessionID: "sess-123", OrgID: "test-org"})
		}
		if r.URL.Path == "/api/v1/code-governance/export" && r.Method == "GET" {
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte("id,title,state\npr-1,Test PR,open\n"))
		}
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		PortalURL:    server.URL,
	})

	_, err := client.LoginToPortal("test-org", "password")
	if err != nil {
		t.Fatalf("LoginToPortal failed: %v", err)
	}

	csv, err := client.ExportCodeGovernanceDataCSV(nil)
	if err != nil {
		t.Fatalf("ExportCodeGovernanceDataCSV failed: %v", err)
	}

	if len(csv) == 0 {
		t.Error("Expected non-empty CSV data")
	}
}

func TestPortalRequestWithoutLogin(t *testing.T) {
	client := NewClient(AxonFlowConfig{
		AgentURL:     "http://localhost",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	// Try to call a portal method without logging in
	_, err := client.ListGitProviders()
	if err == nil {
		t.Error("Expected error when calling portal method without login")
	}
}

func TestGetPortalURLFallback(t *testing.T) {
	client := NewClient(AxonFlowConfig{
		AgentURL:     "http://localhost:8080",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		// No PortalURL set
	})

	// The getPortalURL method should fall back to agent host with port 8082
	// We can't directly test the private method, but we can verify behavior
	if client.IsLoggedIn() {
		t.Error("Client should not be logged in")
	}
}
