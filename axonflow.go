// Package axonflow provides an enterprise-grade Go SDK for the AxonFlow AI governance platform.
// It enables invisible AI governance with production-ready features including retry logic,
// caching, fail-open strategy, and debug mode.
package axonflow

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// AxonFlowConfig represents configuration for the AxonFlow client
type AxonFlowConfig struct {
	AgentURL     string        // Required: AxonFlow Agent URL
	ClientID     string        // Required: Client ID for authentication
	ClientSecret string        // Required: Client secret for authentication
	LicenseKey   string        // Required: AxonFlow license key for agent authentication
	Mode         string        // "production" | "sandbox" (default: "production")
	Debug        bool          // Enable debug logging (default: false)
	Timeout      time.Duration // Request timeout (default: 60s)
	Retry        RetryConfig   // Retry configuration
	Cache        CacheConfig   // Cache configuration
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	Enabled      bool          // Enable retry logic (default: true)
	MaxAttempts  int           // Maximum retry attempts (default: 3)
	InitialDelay time.Duration // Initial delay between retries (default: 1s)
}

// CacheConfig configures caching behavior
type CacheConfig struct {
	Enabled bool          // Enable caching (default: true)
	TTL     time.Duration // Cache TTL (default: 60s)
}

// AxonFlowClient represents the SDK for connecting to AxonFlow platform
type AxonFlowClient struct {
	config     AxonFlowConfig
	httpClient *http.Client
	cache      *cache
}

// ClientRequest represents a request to AxonFlow Agent
type ClientRequest struct {
	Query       string                 `json:"query"`
	UserToken   string                 `json:"user_token"`
	ClientID    string                 `json:"client_id"`
	RequestType string                 `json:"request_type"` // "multi-agent-plan", "sql", "chat", "mcp-query"
	Context     map[string]interface{} `json:"context"`
}

// ClientResponse represents response from AxonFlow Agent
type ClientResponse struct {
	Success     bool                   `json:"success"`
	Data        interface{}            `json:"data,omitempty"`
	Result      string                 `json:"result,omitempty"`     // For multi-agent planning
	PlanID      string                 `json:"plan_id,omitempty"`    // For multi-agent planning
	RequestID   string                 `json:"request_id,omitempty"` // Unique request identifier
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Blocked     bool                   `json:"blocked"`
	BlockReason string                 `json:"block_reason,omitempty"`
	PolicyInfo  *PolicyEvaluationInfo  `json:"policy_info,omitempty"`
}

// PolicyEvaluationInfo contains policy evaluation metadata
type PolicyEvaluationInfo struct {
	PoliciesEvaluated []string `json:"policies_evaluated"`
	StaticChecks      []string `json:"static_checks"`
	ProcessingTime    string   `json:"processing_time"` // Processing time as duration string (e.g., "17.48s")
	TenantID          string   `json:"tenant_id"`
}

// ConnectorMetadata represents information about an MCP connector
type ConnectorMetadata struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Icon         string                 `json:"icon"`
	Tags         []string               `json:"tags"`
	Capabilities []string               `json:"capabilities"`
	ConfigSchema map[string]interface{} `json:"config_schema"`
	Installed    bool                   `json:"installed"`
	InstanceName string                 `json:"instance_name,omitempty"` // Name of installed instance
	Healthy      bool                   `json:"healthy,omitempty"`
}

// ConnectorInstallRequest represents a request to install an MCP connector
type ConnectorInstallRequest struct {
	ConnectorID string                 `json:"connector_id"`
	Name        string                 `json:"name"`
	TenantID    string                 `json:"tenant_id"`
	Options     map[string]interface{} `json:"options"`
	Credentials map[string]string      `json:"credentials"`
}

// ConnectorResponse represents response from an MCP connector query
type ConnectorResponse struct {
	Success bool                   `json:"success"`
	Data    interface{}            `json:"data"`
	Error   string                 `json:"error,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// PlanResponse represents a multi-agent plan generation response
type PlanResponse struct {
	PlanID            string                 `json:"plan_id"`
	Steps             []PlanStep             `json:"steps"`
	Domain            string                 `json:"domain"`
	Complexity        int                    `json:"complexity"`         // Complexity score (1-10)
	Parallel          bool                   `json:"parallel"`           // Whether steps can run in parallel
	EstimatedDuration string                 `json:"estimated_duration"` // Estimated execution time
	Metadata          map[string]interface{} `json:"metadata"`
}

// PlanStep represents a single step in a multi-agent plan
type PlanStep struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Description   string                 `json:"description"`
	Dependencies  []string               `json:"dependencies"` // IDs of steps this depends on
	Agent         string                 `json:"agent"`        // Agent responsible for execution
	Parameters    map[string]interface{} `json:"parameters"`
	EstimatedTime string                 `json:"estimated_time"` // Estimated execution time for this step
}

// PlanExecutionResponse represents the result of plan execution
type PlanExecutionResponse struct {
	PlanID                 string       `json:"plan_id"`
	Status                 string       `json:"status"` // "running", "completed", "failed", "partial"
	Result                 string       `json:"result,omitempty"`
	StepResults            []StepResult `json:"step_results,omitempty"`
	Error                  string       `json:"error,omitempty"`
	Duration               string       `json:"duration,omitempty"`
	CompletedSteps         int          `json:"completed_steps"`                    // Number of completed steps
	TotalSteps             int          `json:"total_steps"`                        // Total number of steps
	CurrentStep            string       `json:"current_step,omitempty"`             // Currently executing step
	EstimatedTimeRemaining string       `json:"estimated_time_remaining,omitempty"` // For in-progress plans
}

// StepResult represents the result of a single plan step execution
type StepResult struct {
	StepID   string      `json:"step_id"`
	StepName string      `json:"step_name"`
	Status   string      `json:"status"` // "pending", "running", "completed", "failed"
	Result   interface{} `json:"result,omitempty"`
	Error    string      `json:"error,omitempty"`
	Duration string      `json:"duration,omitempty"`
}

// Cache entry
type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

// Simple in-memory cache
type cache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	ttl     time.Duration
}

func newCache(ttl time.Duration) *cache {
	c := &cache{
		entries: make(map[string]*cacheEntry),
		ttl:     ttl,
	}
	// Start cleanup goroutine
	go c.cleanup()
	return c
}

func (c *cache) get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiration) {
		return nil, false
	}

	return entry.value, true
}

func (c *cache) set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &cacheEntry{
		value:      value,
		expiration: time.Now().Add(c.ttl),
	}
}

func (c *cache) cleanup() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.expiration) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// NewClient creates a new AxonFlow client with the given configuration
func NewClient(config AxonFlowConfig) *AxonFlowClient {
	// Check if running in self-hosted mode (localhost)
	isLocalhost := strings.Contains(config.AgentURL, "localhost") || strings.Contains(config.AgentURL, "127.0.0.1")

	// Set defaults
	if config.Mode == "" {
		if isLocalhost {
			config.Mode = "sandbox"
		} else {
			config.Mode = "production"
		}
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.Retry.InitialDelay == 0 {
		config.Retry.InitialDelay = 1 * time.Second
	}
	if config.Retry.MaxAttempts == 0 {
		config.Retry.MaxAttempts = 3
		config.Retry.Enabled = true
	}
	if config.Cache.TTL == 0 {
		config.Cache.TTL = 60 * time.Second
		config.Cache.Enabled = true
	}

	// Configure TLS
	tlsConfig := &tls.Config{}
	if os.Getenv("NODE_TLS_REJECT_UNAUTHORIZED") == "0" {
		tlsConfig.InsecureSkipVerify = true
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &AxonFlowClient{
		config: config,
		httpClient: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
	}

	if config.Cache.Enabled {
		client.cache = newCache(config.Cache.TTL)
	}

	if config.Debug {
		authMethod := "license-key"
		if config.LicenseKey == "" && config.ClientSecret != "" {
			authMethod = "client-secret"
		} else if isLocalhost && config.LicenseKey == "" && config.ClientSecret == "" {
			authMethod = "self-hosted (no auth)"
		}
		log.Printf("[AxonFlow] Client initialized - Mode: %s, Endpoint: %s, Auth: %s", config.Mode, config.AgentURL, authMethod)
	}

	return client
}

// NewClientSimple creates a client with simple parameters (backward compatible)
func NewClientSimple(agentURL, clientID, clientSecret string) *AxonFlowClient {
	return NewClient(AxonFlowConfig{
		AgentURL:     agentURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
}

// Sandbox creates a client in sandbox mode for testing
func Sandbox(apiKey string) *AxonFlowClient {
	if apiKey == "" {
		apiKey = "demo-key"
	}

	return NewClient(AxonFlowConfig{
		AgentURL:     "https://staging-eu.getaxonflow.com",
		ClientID:     apiKey,
		ClientSecret: apiKey,
		Mode:         "sandbox",
		Debug:        true,
	})
}

// ExecuteQuery sends a query through AxonFlow platform with policy enforcement
func (c *AxonFlowClient) ExecuteQuery(userToken, query, requestType string, context map[string]interface{}) (*ClientResponse, error) {
	// Generate cache key
	cacheKey := fmt.Sprintf("%s:%s:%s", requestType, query, userToken)

	// Check cache if enabled
	if c.cache != nil {
		if cached, found := c.cache.get(cacheKey); found {
			if c.config.Debug {
				log.Printf("[AxonFlow] Cache hit for query: %s", query[:min(50, len(query))])
			}
			return cached.(*ClientResponse), nil
		}
	}

	req := ClientRequest{
		Query:       query,
		UserToken:   userToken,
		ClientID:    c.config.ClientID,
		RequestType: requestType,
		Context:     context,
	}

	var resp *ClientResponse
	var err error

	// Execute with retry if enabled
	if c.config.Retry.Enabled {
		resp, err = c.executeWithRetry(req)
	} else {
		resp, err = c.executeRequest(req)
	}

	// Handle fail-open in production mode
	if err != nil && c.config.Mode == "production" && c.isAxonFlowError(err) {
		if c.config.Debug {
			log.Printf("[AxonFlow] AxonFlow unavailable, failing open: %v", err)
		}
		// Return a success response indicating the request was allowed through
		return &ClientResponse{
			Success: true,
			Data:    nil,
			Error:   fmt.Sprintf("AxonFlow unavailable (fail-open): %v", err),
		}, nil
	}

	if err != nil {
		return nil, err
	}

	// Cache successful responses
	if c.cache != nil && resp.Success {
		c.cache.set(cacheKey, resp)
	}

	return resp, nil
}

// executeWithRetry executes a request with exponential backoff retry
func (c *AxonFlowClient) executeWithRetry(req ClientRequest) (*ClientResponse, error) {
	var lastErr error

	for attempt := 0; attempt < c.config.Retry.MaxAttempts; attempt++ {
		if attempt > 0 {
			// Exponential backoff: delay * 2^(attempt-1)
			delay := time.Duration(float64(c.config.Retry.InitialDelay) * math.Pow(2, float64(attempt-1)))
			if c.config.Debug {
				log.Printf("[AxonFlow] Retry attempt %d/%d after %v", attempt+1, c.config.Retry.MaxAttempts, delay)
			}
			time.Sleep(delay)
		}

		resp, err := c.executeRequest(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on client errors (4xx)
		if httpErr, ok := err.(*httpError); ok && httpErr.statusCode >= 400 && httpErr.statusCode < 500 {
			if c.config.Debug {
				log.Printf("[AxonFlow] Client error (4xx), not retrying: %v", err)
			}
			break
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.config.Retry.MaxAttempts, lastErr)
}

// httpError represents an HTTP error with status code
type httpError struct {
	statusCode int
	message    string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.statusCode, e.message)
}

// executeRequest executes a single request without retry
func (c *AxonFlowClient) executeRequest(req ClientRequest) (*ClientResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.config.AgentURL+"/api/request", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Skip auth headers for localhost (self-hosted mode)
	isLocalhost := strings.Contains(c.config.AgentURL, "localhost") || strings.Contains(c.config.AgentURL, "127.0.0.1")
	if !isLocalhost {
		if c.config.ClientSecret != "" {
			httpReq.Header.Set("X-Client-Secret", c.config.ClientSecret)
		}
		if c.config.LicenseKey != "" {
			httpReq.Header.Set("X-License-Key", c.config.LicenseKey)
		}
	}

	if c.config.Debug {
		log.Printf("[AxonFlow] Sending request - Type: %s, Query: %s", req.RequestType, req.Query[:min(50, len(req.Query))])
	}

	startTime := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// [DEBUG] Log raw response body before unmarshaling
	log.Printf("[SDK-DEBUG] Raw response body size: %d bytes", len(body))
	if len(body) > 0 && len(body) <= 500 {
		log.Printf("[SDK-DEBUG] Raw response body (full): %s", string(body))
	} else if len(body) > 500 {
		log.Printf("[SDK-DEBUG] Raw response body (first 500 chars): %s...", string(body[:500]))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &httpError{
			statusCode: resp.StatusCode,
			message:    string(body),
		}
	}

	var clientResp ClientResponse
	if err := json.Unmarshal(body, &clientResp); err != nil {
		log.Printf("[SDK-DEBUG] Unmarshal error: %v", err)
		log.Printf("[SDK-DEBUG] Body that failed to unmarshal: %s", string(body))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for nested errors in the Data field
	// When orchestrator fails, agent wraps error as: {"success":true, "data":{"success":false, "error":"..."}}
	if clientResp.Data != nil {
		if dataMap, ok := clientResp.Data.(map[string]interface{}); ok {
			// Check if data contains nested success field
			if dataSuccess, hasSuccess := dataMap["success"].(bool); hasSuccess && !dataSuccess {
				// Orchestrator execution failed - extract error message
				if errorMsg, hasError := dataMap["error"].(string); hasError {
					log.Printf("[SDK-DEBUG] Detected orchestrator failure in data.error: %s", errorMsg)
					// Surface the error by setting the Error field and marking success as false
					clientResp.Error = errorMsg
					clientResp.Success = false
				}
			}
			// Also check if data.result exists and use it if Result is empty
			if clientResp.Result == "" {
				if dataResult, hasResult := dataMap["result"].(string); hasResult && dataResult != "" {
					log.Printf("[SDK-DEBUG] Using data.result field (length: %d)", len(dataResult))
					clientResp.Result = dataResult
				}
			}
			// Check if data.plan_id exists and use it if PlanID is empty
			if clientResp.PlanID == "" {
				if dataPlanID, hasPlanID := dataMap["plan_id"].(string); hasPlanID && dataPlanID != "" {
					log.Printf("[SDK-DEBUG] Using data.plan_id field: %s", dataPlanID)
					clientResp.PlanID = dataPlanID
				}
			}
			// Check if data.metadata exists and use it if Metadata is empty
			if clientResp.Metadata == nil {
				if dataMetadata, hasMetadata := dataMap["metadata"].(map[string]interface{}); hasMetadata {
					log.Printf("[SDK-DEBUG] Using data.metadata field")
					clientResp.Metadata = dataMetadata
				}
			}
		}
	}

	// [DEBUG] Log unmarshaled response details
	log.Printf("[SDK-DEBUG] Unmarshaled - Success: %v, Result length: %d, PlanID: %s",
		clientResp.Success, len(clientResp.Result), clientResp.PlanID)
	if len(clientResp.Result) > 0 {
		if len(clientResp.Result) <= 100 {
			log.Printf("[SDK-DEBUG] Result (full): %s", clientResp.Result)
		} else {
			log.Printf("[SDK-DEBUG] Result (first 100 chars): %s...", clientResp.Result[:100])
		}
	} else {
		log.Printf("[SDK-DEBUG] Result is empty!")
	}
	log.Printf("[SDK-DEBUG] Metadata keys: %v", getMetadataKeys(clientResp.Metadata))

	// If we detected an error in the data field, log it prominently
	if clientResp.Error != "" {
		log.Printf("[SDK-DEBUG] Error field set: %s", clientResp.Error)
	}

	if c.config.Debug {
		log.Printf("[AxonFlow] Response received - Success: %v, Duration: %v", clientResp.Success, duration)
	}

	return &clientResp, nil
}

// isAxonFlowError checks if an error is from AxonFlow (vs the AI provider)
func (c *AxonFlowClient) isAxonFlowError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "AxonFlow") ||
		strings.Contains(errMsg, "governance") ||
		strings.Contains(errMsg, "request failed") ||
		strings.Contains(errMsg, "connection refused")
}

// HealthCheck checks if AxonFlow Agent is healthy
func (c *AxonFlowClient) HealthCheck() error {
	resp, err := c.httpClient.Get(c.config.AgentURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("agent not healthy: status %d", resp.StatusCode)
	}

	if c.config.Debug {
		log.Println("[AxonFlow] Health check passed")
	}

	return nil
}

// getMetadataKeys returns the keys from a metadata map for debugging
func getMetadataKeys(metadata map[string]interface{}) []string {
	if metadata == nil {
		return []string{}
	}
	keys := make([]string, 0, len(metadata))
	for k := range metadata {
		keys = append(keys, k)
	}
	return keys
}

// ListConnectors returns all available MCP connectors from the marketplace
func (c *AxonFlowClient) ListConnectors() ([]ConnectorMetadata, error) {
	resp, err := c.httpClient.Get(c.config.AgentURL + "/api/connectors")
	if err != nil {
		return nil, fmt.Errorf("failed to list connectors: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list connectors failed: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var connectors []ConnectorMetadata
	if err := json.NewDecoder(resp.Body).Decode(&connectors); err != nil {
		return nil, fmt.Errorf("failed to decode connectors: %w", err)
	}

	if c.config.Debug {
		log.Printf("[AxonFlow] Listed %d connectors", len(connectors))
	}

	return connectors, nil
}

// InstallConnector installs an MCP connector from the marketplace
func (c *AxonFlowClient) InstallConnector(req ConnectorInstallRequest) error {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal install request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.config.AgentURL+"/api/connectors/install", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create install request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Skip auth headers for localhost (self-hosted mode)
	isLocalhost := strings.Contains(c.config.AgentURL, "localhost") || strings.Contains(c.config.AgentURL, "127.0.0.1")
	if !isLocalhost {
		if c.config.ClientSecret != "" {
			httpReq.Header.Set("X-Client-Secret", c.config.ClientSecret)
		}
		if c.config.LicenseKey != "" {
			httpReq.Header.Set("X-License-Key", c.config.LicenseKey)
		}
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("install request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("install failed: HTTP %d: %s", resp.StatusCode, string(body))
	}

	if c.config.Debug {
		log.Printf("[AxonFlow] Connector installed: %s", req.Name)
	}

	return nil
}

// QueryConnector executes a query against an installed MCP connector
func (c *AxonFlowClient) QueryConnector(userToken, connectorName, query string, params map[string]interface{}) (*ConnectorResponse, error) {
	context := map[string]interface{}{
		"connector": connectorName,
		"params":    params,
	}

	resp, err := c.ExecuteQuery(userToken, query, "mcp-query", context)
	if err != nil {
		return nil, err
	}

	connResp := &ConnectorResponse{
		Success: resp.Success,
		Data:    resp.Data,
		Error:   resp.Error,
		Meta:    resp.Metadata,
	}

	return connResp, nil
}

// GeneratePlan creates a multi-agent execution plan from a natural language query
func (c *AxonFlowClient) GeneratePlan(query string, domain string) (*PlanResponse, error) {
	context := map[string]interface{}{}
	if domain != "" {
		context["domain"] = domain
	}

	resp, err := c.ExecuteQuery("", query, "multi-agent-plan", context)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("plan generation failed: %s", resp.Error)
	}

	// Parse plan from response
	planData, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected plan response format")
	}

	// Convert to PlanResponse
	planBytes, _ := json.Marshal(planData)
	var plan PlanResponse
	if err := json.Unmarshal(planBytes, &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan: %w", err)
	}

	plan.PlanID = resp.PlanID

	if c.config.Debug {
		log.Printf("[AxonFlow] Plan generated: %s (%d steps)", plan.PlanID, len(plan.Steps))
	}

	return &plan, nil
}

// ExecutePlan executes a previously generated multi-agent plan
func (c *AxonFlowClient) ExecutePlan(planID string) (*PlanExecutionResponse, error) {
	context := map[string]interface{}{
		"plan_id": planID,
	}

	resp, err := c.ExecuteQuery("", "", "execute-plan", context)
	if err != nil {
		return nil, err
	}

	execResp := &PlanExecutionResponse{
		PlanID: planID,
		Status: "completed",
		Result: resp.Result,
		Error:  resp.Error,
	}

	if resp.Metadata != nil {
		if duration, ok := resp.Metadata["duration"].(string); ok {
			execResp.Duration = duration
		}
		if stepResults, ok := resp.Metadata["step_results"].([]interface{}); ok {
			// Convert to StepResult slice
			for _, sr := range stepResults {
				if srMap, ok := sr.(map[string]interface{}); ok {
					stepResult := StepResult{}
					if id, ok := srMap["step_id"].(string); ok {
						stepResult.StepID = id
					}
					if name, ok := srMap["step_name"].(string); ok {
						stepResult.StepName = name
					}
					if status, ok := srMap["status"].(string); ok {
						stepResult.Status = status
					}
					if result, ok := srMap["result"]; ok {
						stepResult.Result = result
					}
					if errStr, ok := srMap["error"].(string); ok {
						stepResult.Error = errStr
					}
					if dur, ok := srMap["duration"].(string); ok {
						stepResult.Duration = dur
					}
					execResp.StepResults = append(execResp.StepResults, stepResult)
				}
			}
		}
		if completed, ok := resp.Metadata["completed_steps"].(float64); ok {
			execResp.CompletedSteps = int(completed)
		}
		if total, ok := resp.Metadata["total_steps"].(float64); ok {
			execResp.TotalSteps = int(total)
		}
	}

	if !resp.Success {
		execResp.Status = "failed"
	}

	if c.config.Debug {
		log.Printf("[AxonFlow] Plan executed: %s - Status: %s", planID, execResp.Status)
	}

	return execResp, nil
}

// GetPlanStatus retrieves the status of a running or completed plan
func (c *AxonFlowClient) GetPlanStatus(planID string) (*PlanExecutionResponse, error) {
	resp, err := c.httpClient.Get(c.config.AgentURL + "/api/plans/" + planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get plan status failed: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var status PlanExecutionResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode plan status: %w", err)
	}

	return &status, nil
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
