// Copyright 2025 AxonFlow
// SPDX-License-Identifier: BUSL-1.1
//
// selfhosted_auth_headers_test.go - Auth header verification tests
//
// These tests verify that auth headers are:
// - Sent when credentials are provided
// - NOT sent when credentials are not provided (community/self-hosted mode)
//
// Run with:
//   go test -v -run TestAuthHeaders

package axonflow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ============================================================
// AUTH HEADERS TESTS - Community Mode (no credentials)
// ============================================================

// TestAuthHeaders_NotSentWithoutCredentials verifies auth headers
// are NOT sent when credentials are not configured (community/self-hosted mode)
func TestAuthHeaders_NotSentWithoutCredentials(t *testing.T) {
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

	// Create client WITHOUT credentials
	client := NewClient(AxonFlowConfig{
		AgentURL: server.URL,
		// No ClientID/ClientSecret/LicenseKey
		Cache: CacheConfig{Enabled: false},
	})

	_, _ = client.ExecuteQuery("user", "query", "chat", nil)

	// Auth headers should NOT be set when no credentials are provided
	if receivedAuthHeader != "" {
		t.Errorf("Expected no X-Client-Secret header without credentials, got '%s'", receivedAuthHeader)
	}
	if receivedLicenseHeader != "" {
		t.Errorf("Expected no X-License-Key header without credentials, got '%s'", receivedLicenseHeader)
	}

	t.Log("✅ Auth headers correctly NOT sent in community mode (no credentials)")
}

// ============================================================
// AUTH HEADERS TESTS - Enterprise Mode (with credentials)
// ============================================================

// TestAuthHeaders_SentWithCredentials verifies auth headers
// ARE sent when credentials are configured (enterprise mode)
func TestAuthHeaders_SentWithCredentials(t *testing.T) {
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

	// Create client WITH credentials
	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test",
		ClientSecret: "secret",
		LicenseKey:   "license-key",
		Cache:        CacheConfig{Enabled: false},
	})

	_, _ = client.ExecuteQuery("user", "query", "chat", nil)

	// Auth headers SHOULD be set when credentials are provided
	if receivedAuthHeader != "secret" {
		t.Errorf("Expected X-Client-Secret 'secret', got '%s'", receivedAuthHeader)
	}
	if receivedLicenseHeader != "license-key" {
		t.Errorf("Expected X-License-Key 'license-key', got '%s'", receivedLicenseHeader)
	}

	t.Log("✅ Auth headers correctly sent in enterprise mode (with credentials)")
}

// TestAuthHeaders_OnlyLicenseKey verifies auth headers work with just LicenseKey
func TestAuthHeaders_OnlyLicenseKey(t *testing.T) {
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

	// Create client with ONLY LicenseKey
	client := NewClient(AxonFlowConfig{
		AgentURL:   server.URL,
		LicenseKey: "license-key",
		Cache:      CacheConfig{Enabled: false},
	})

	_, _ = client.ExecuteQuery("user", "query", "chat", nil)

	// Only LicenseKey should be sent, not ClientSecret
	if receivedAuthHeader != "" {
		t.Errorf("Expected no X-Client-Secret with only LicenseKey, got '%s'", receivedAuthHeader)
	}
	if receivedLicenseHeader != "license-key" {
		t.Errorf("Expected X-License-Key 'license-key', got '%s'", receivedLicenseHeader)
	}

	t.Log("✅ LicenseKey header correctly sent when only LicenseKey is configured")
}

// TestAuthHeaders_OnlyClientSecret verifies auth headers work with just ClientSecret
func TestAuthHeaders_OnlyClientSecret(t *testing.T) {
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

	// Create client with ONLY ClientSecret
	client := NewClient(AxonFlowConfig{
		AgentURL:     server.URL,
		ClientID:     "test",
		ClientSecret: "secret",
		Cache:        CacheConfig{Enabled: false},
	})

	_, _ = client.ExecuteQuery("user", "query", "chat", nil)

	// Only ClientSecret should be sent, not LicenseKey
	if receivedAuthHeader != "secret" {
		t.Errorf("Expected X-Client-Secret 'secret', got '%s'", receivedAuthHeader)
	}
	if receivedLicenseHeader != "" {
		t.Errorf("Expected no X-License-Key with only ClientSecret, got '%s'", receivedLicenseHeader)
	}

	t.Log("✅ ClientSecret header correctly sent when only ClientSecret is configured")
}
