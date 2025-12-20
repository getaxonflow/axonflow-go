// Copyright 2025 AxonFlow
// SPDX-License-Identifier: BUSL-1.1
//
// selfhosted_auth_headers_test.go - Auth header verification tests
//
// These tests verify that auth headers are NOT sent for localhost endpoints
// (self-hosted mode). They use httptest and don't require a running agent.
//
// Run with:
//   go test -v -run TestZeroConfig_AuthHeaders

package axonflow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ============================================================
// 7. AUTH HEADERS NOT SENT FOR LOCALHOST
// ============================================================

// TestZeroConfig_AuthHeaders_NotSentForLocalhost verifies auth headers
// are NOT sent for localhost endpoints (self-hosted mode)
func TestZeroConfig_AuthHeaders_NotSentForLocalhost(t *testing.T) {
	receivedAuthHeader := ""
	receivedLicenseHeader := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("X-Client-Secret")
		receivedLicenseHeader = r.Header.Get("X-License-Key")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    map[string]string{"answer": "test"},
		})
	}))
	defer server.Close()

	// httptest server URL contains 127.0.0.1, so auth should be skipped
	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test",
		ClientSecret: "secret",
		LicenseKey:   "license-key",
		Cache:        CacheConfig{Enabled: false},
	})

	_, _ = client.ExecuteQuery("user", "query", "chat", nil)

	// Auth headers should NOT be set for localhost/127.0.0.1 (self-hosted mode)
	if receivedAuthHeader != "" {
		t.Errorf("Expected no X-Client-Secret header for localhost, got '%s'", receivedAuthHeader)
	}
	if receivedLicenseHeader != "" {
		t.Errorf("Expected no X-License-Key header for localhost, got '%s'", receivedLicenseHeader)
	}

	t.Log("✅ Auth headers correctly NOT sent for localhost")
}

// TestZeroConfig_AuthHeaders_PreCheckNoAuth verifies pre-check doesn't
// send auth headers for localhost
func TestZeroConfig_AuthHeaders_PreCheckNoAuth(t *testing.T) {
	receivedAuthHeader := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("X-Client-Secret")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"context_id": "ctx_test_123",
			"approved":   true,
			"expires_at": time.Now().Add(5 * time.Minute).Format(time.RFC3339),
		})
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test",
		ClientSecret: "secret",
		Cache:        CacheConfig{Enabled: false},
	})

	result, err := client.GetPolicyApprovedContext("", "test query", nil, nil)
	if err != nil {
		t.Fatalf("GetPolicyApprovedContext failed: %v", err)
	}

	if result.ContextID != "ctx_test_123" {
		t.Errorf("Expected context_id 'ctx_test_123', got '%s'", result.ContextID)
	}

	if receivedAuthHeader != "" {
		t.Errorf("Expected no X-Client-Secret header for localhost, got '%s'", receivedAuthHeader)
	}

	t.Log("✅ Pre-check auth headers correctly NOT sent for localhost")
}

// TestZeroConfig_AuthHeaders_AuditNoAuth verifies audit doesn't
// send auth headers for localhost
func TestZeroConfig_AuthHeaders_AuditNoAuth(t *testing.T) {
	receivedAuthHeader := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("X-Client-Secret")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"audit_id": "audit_test_123",
		})
	}))
	defer server.Close()

	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test",
		ClientSecret: "secret",
		Cache:        CacheConfig{Enabled: false},
	})

	result, err := client.AuditLLMCall(
		"ctx_test",
		"response summary",
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

	if receivedAuthHeader != "" {
		t.Errorf("Expected no X-Client-Secret header for localhost, got '%s'", receivedAuthHeader)
	}

	t.Log("✅ Audit auth headers correctly NOT sent for localhost")
}
