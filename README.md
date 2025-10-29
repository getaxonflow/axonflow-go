# AxonFlow SDK for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/getaxonflow/axonflow-go.svg)](https://pkg.go.dev/github.com/getaxonflow/axonflow-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/getaxonflow/axonflow-go)](https://goreportcard.com/report/github.com/getaxonflow/axonflow-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Enterprise-grade Go SDK for AxonFlow AI governance platform. Add invisible AI governance to your applications with production-ready features including retry logic, caching, fail-open strategy, and debug mode.

## Installation

```bash
go get github.com/getaxonflow/axonflow-go
```

## Quick Start

### Basic Usage (License-Based Auth)

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/getaxonflow/axonflow-go"
)

func main() {
    // Simple initialization with license key
    client := axonflow.NewClient(axonflow.AxonFlowConfig{
        AgentURL:   "https://staging-eu.getaxonflow.com",
        LicenseKey: os.Getenv("AXONFLOW_LICENSE_KEY"),
    })

    // Execute a governed query
    resp, err := client.ExecuteQuery(
        "user-token",
        "What is the capital of France?",
        "chat",
        map[string]interface{}{},
    )

    if err != nil {
        log.Fatalf("Query failed: %v", err)
    }

    if resp.Blocked {
        log.Printf("Request blocked: %s", resp.BlockReason)
        return
    }

    fmt.Printf("Result: %s\n", resp.Data)
}
```

### Advanced Configuration

```go
import (
    "time"
    "os"
    "github.com/getaxonflow/axonflow-go"
)

// Full configuration with all features
client := axonflow.NewClient(axonflow.AxonFlowConfig{
    AgentURL:   "https://staging-eu.getaxonflow.com",
    LicenseKey: os.Getenv("AXONFLOW_LICENSE_KEY"),
    Mode:       "production",  // or "sandbox"
    Debug:      true,          // Enable debug logging
    Timeout:    60 * time.Second,

    // Retry configuration (exponential backoff)
    Retry: axonflow.RetryConfig{
        Enabled:      true,
        MaxAttempts:  3,
        InitialDelay: 1 * time.Second,
    },

    // Cache configuration (in-memory with TTL)
    Cache: axonflow.CacheConfig{
        Enabled: true,
        TTL:     60 * time.Second,
    },
})
```

### Legacy Authentication (Deprecated)

> **⚠️ Deprecated:** `ClientID` and `ClientSecret` authentication is deprecated. Please migrate to license-based authentication using `LicenseKey`.

```go
// Legacy method (still supported for backward compatibility)
client := axonflow.NewClientSimple(
    "https://staging-eu.getaxonflow.com",
    "your-client-id",
    "your-client-secret",
)
```

### Sandbox Mode (Testing)

```go
// Quick sandbox client for testing
client := axonflow.Sandbox("demo-key")

resp, err := client.ExecuteQuery(
    "",
    "Test query with sensitive data: SSN 123-45-6789",
    "chat",
    map[string]interface{}{},
)

// In sandbox, this will be blocked/redacted
if resp.Blocked {
    fmt.Printf("Blocked: %s\n", resp.BlockReason)
}
```

## Features

### ✅ Retry Logic with Exponential Backoff

Automatic retry on transient failures with exponential backoff:

```go
client := axonflow.NewClient(axonflow.AxonFlowConfig{
    AgentURL: "https://staging-eu.getaxonflow.com",
    ClientID: "your-client-id",
    ClientSecret: "your-secret",
    Retry: axonflow.RetryConfig{
        Enabled:      true,
        MaxAttempts:  3,           // Retry up to 3 times
        InitialDelay: 1 * time.Second,  // 1s, 2s, 4s backoff
    },
})

// Automatically retries on 5xx errors or network failures
resp, err := client.ExecuteQuery(...)
```

### ✅ In-Memory Caching with TTL

Reduce latency and load with intelligent caching:

```go
client := axonflow.NewClient(axonflow.AxonFlowConfig{
    AgentURL: "https://staging-eu.getaxonflow.com",
    ClientID: "your-client-id",
    ClientSecret: "your-secret",
    Cache: axonflow.CacheConfig{
        Enabled: true,
        TTL:     60 * time.Second,  // Cache for 60 seconds
    },
})

// First call: hits AxonFlow
resp1, _ := client.ExecuteQuery("token", "query", "chat", nil)

// Second call (within 60s): served from cache
resp2, _ := client.ExecuteQuery("token", "query", "chat", nil)
```

### ✅ Fail-Open Strategy (Production Mode)

Never block your users if AxonFlow is unavailable:

```go
client := axonflow.NewClient(axonflow.AxonFlowConfig{
    AgentURL: "https://staging-eu.getaxonflow.com",
    ClientID: "your-client-id",
    ClientSecret: "your-secret",
    Mode:     "production",  // Fail-open in production
    Debug:    true,
})

// If AxonFlow is unavailable, request proceeds with warning
resp, err := client.ExecuteQuery(...)
// err == nil, resp.Success == true, resp.Error contains warning
```

## MCP Connector Marketplace

Integrate with external data sources using AxonFlow's MCP (Model Context Protocol) connectors:

### List Available Connectors

```go
connectors, err := client.ListConnectors()
if err != nil {
    log.Fatalf("Failed to list connectors: %v", err)
}

for _, conn := range connectors {
    fmt.Printf("Connector: %s (%s)\n", conn.Name, conn.Type)
    fmt.Printf("  Description: %s\n", conn.Description)
    fmt.Printf("  Installed: %v\n", conn.Installed)
}
```

### Install a Connector

```go
err := client.InstallConnector(axonflow.ConnectorInstallRequest{
    ConnectorID: "amadeus-travel",
    Name:        "amadeus-prod",
    TenantID:    "your-tenant-id",
    Options: map[string]interface{}{
        "environment": "production",
    },
    Credentials: map[string]string{
        "api_key":    "your-amadeus-key",
        "api_secret": "your-amadeus-secret",
    },
})

if err != nil {
    log.Fatalf("Failed to install connector: %v", err)
}

fmt.Println("Connector installed successfully!")
```

### Query a Connector

```go
// Query the Amadeus connector for flight information
resp, err := client.QueryConnector(
    "amadeus-prod",
    "Find flights from Paris to Amsterdam on Dec 15",
    map[string]interface{}{
        "origin":      "CDG",
        "destination": "AMS",
        "date":        "2025-12-15",
    },
)

if err != nil {
    log.Fatalf("Connector query failed: %v", err)
}

if resp.Success {
    fmt.Printf("Flight data: %v\n", resp.Data)
} else {
    fmt.Printf("Query failed: %s\n", resp.Error)
}
```

## Multi-Agent Planning (MAP)

Generate and execute complex multi-step plans using AI agent orchestration:

### Generate a Plan

```go
// Generate a travel planning workflow
plan, err := client.GeneratePlan(
    "Plan a 3-day trip to Paris with moderate budget",
    "travel",  // Domain hint (optional)
)

if err != nil {
    log.Fatalf("Plan generation failed: %v", err)
}

fmt.Printf("Generated plan %s with %d steps\n", plan.PlanID, len(plan.Steps))
fmt.Printf("Complexity: %d, Parallel: %v\n", plan.Complexity, plan.Parallel)

for i, step := range plan.Steps {
    fmt.Printf("  Step %d: %s (%s)\n", i+1, step.Name, step.Type)
    fmt.Printf("    Description: %s\n", step.Description)
    fmt.Printf("    Agent: %s\n", step.Agent)
}
```

### Execute a Plan

```go
// Execute the generated plan
execResp, err := client.ExecutePlan(plan.PlanID)
if err != nil {
    log.Fatalf("Plan execution failed: %v", err)
}

fmt.Printf("Plan Status: %s\n", execResp.Status)
fmt.Printf("Duration: %s\n", execResp.Duration)

if execResp.Status == "completed" {
    fmt.Printf("Result:\n%s\n", execResp.Result)
} else if execResp.Status == "failed" {
    fmt.Printf("Error: %s\n", execResp.Error)
}
```

### Check Plan Status

```go
// For long-running plans, check status periodically
status, err := client.GetPlanStatus(plan.PlanID)
if err != nil {
    log.Fatalf("Failed to get plan status: %v", err)
}

fmt.Printf("Plan Status: %s\n", status.Status)
```

## Health Check

Check if AxonFlow Agent is available:

```go
err := client.HealthCheck()
if err != nil {
    log.Printf("AxonFlow Agent is unhealthy: %v", err)
} else {
    log.Println("AxonFlow Agent is healthy")
}
```

## VPC Private Endpoint (Low-Latency)

For applications running in AWS VPC, use the private endpoint for sub-10ms latency:

```go
client := axonflow.NewClient(axonflow.AxonFlowConfig{
    AgentURL:     "https://10.0.2.67:8443",  // VPC private endpoint
    ClientID:     "your-client-id",
    ClientSecret: "your-secret",
    Mode:         "production",
})

// Enjoy sub-10ms P99 latency vs ~100ms over public internet
```

**Performance:**
- Public endpoint: ~100ms (internet routing)
- VPC private endpoint: <10ms P99 (intra-VPC routing)

## Error Handling

```go
resp, err := client.ExecuteQuery(...)
if err != nil {
    // Network errors, timeouts, or AxonFlow unavailability
    log.Printf("Request failed: %v", err)
    return
}

if resp.Blocked {
    // Policy violation - request blocked by governance rules
    log.Printf("Request blocked: %s", resp.BlockReason)
    log.Printf("Policies evaluated: %v", resp.PolicyInfo.PoliciesEvaluated)
    return
}

if !resp.Success {
    // Request succeeded but returned error from downstream
    log.Printf("Query failed: %s", resp.Error)
    return
}

// Success - use resp.Data or resp.Result
fmt.Printf("Result: %v\n", resp.Data)
```

## Production Best Practices

1. **Environment Variables**: Never hardcode credentials
   ```go
   import "os"

   client := axonflow.NewClient(axonflow.AxonFlowConfig{
       AgentURL:   os.Getenv("AXONFLOW_AGENT_URL"),
       LicenseKey: os.Getenv("AXONFLOW_LICENSE_KEY"),
   })
   ```

2. **Fail-Open in Production**: Use `Mode: "production"` to fail-open if AxonFlow is unavailable

3. **Enable Caching**: Reduce latency for repeated queries

4. **Enable Retry**: Handle transient failures automatically

5. **Debug in Development**: Use `Debug: true` during development, disable in production

6. **Health Checks**: Monitor AxonFlow availability with periodic health checks

7. **Secure Storage**: Store license keys in environment variables or secrets management systems (AWS Secrets Manager, HashiCorp Vault, etc.)

## Configuration Reference

### AxonFlowConfig

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `AgentURL` | `string` | Required | AxonFlow Agent endpoint URL |
| `LicenseKey` | `string` | **Recommended** | License key for authentication |
| `ClientID` | `string` | Deprecated | ⚠️ Legacy: Client ID (use LicenseKey instead) |
| `ClientSecret` | `string` | Deprecated | ⚠️ Legacy: Client secret (use LicenseKey instead) |
| `Mode` | `string` | `"production"` | `"production"` or `"sandbox"` |
| `Debug` | `bool` | `false` | Enable debug logging |
| `Timeout` | `time.Duration` | `60s` | Request timeout |
| `Retry.Enabled` | `bool` | `true` | Enable retry logic |
| `Retry.MaxAttempts` | `int` | `3` | Maximum retry attempts |
| `Retry.InitialDelay` | `time.Duration` | `1s` | Initial retry delay (exponential backoff) |
| `Cache.Enabled` | `bool` | `true` | Enable caching |
| `Cache.TTL` | `time.Duration` | `60s` | Cache time-to-live |

## Migration Guide

### Migrating from ClientID/ClientSecret to License Key

If you're currently using `ClientID` and `ClientSecret`, migrate to license-based authentication:

**Before:**
```go
client := axonflow.NewClient(axonflow.AxonFlowConfig{
    AgentURL:     "https://staging-eu.getaxonflow.com",
    ClientID:     os.Getenv("AXONFLOW_CLIENT_ID"),
    ClientSecret: os.Getenv("AXONFLOW_CLIENT_SECRET"),
})
```

**After:**
```go
client := axonflow.NewClient(axonflow.AxonFlowConfig{
    AgentURL:   "https://staging-eu.getaxonflow.com",
    LicenseKey: os.Getenv("AXONFLOW_LICENSE_KEY"),
})
```

**How to get a license key:**
1. Contact AxonFlow support at [dev@getaxonflow.com](mailto:dev@getaxonflow.com)
2. License keys are provided as part of your AxonFlow subscription
3. Store keys securely in environment variables or secrets management systems

## Examples

See the [examples](examples/) directory for complete working examples:

- [Basic Usage](examples/basic/main.go)
- [MCP Connectors](examples/connectors/main.go)
- [Multi-Agent Planning](examples/planning/main.go)

## Support

- **Documentation**: https://docs.getaxonflow.com
- **npm SDK**: https://www.npmjs.com/package/@axonflow/sdk
- **Issues**: https://github.com/getaxonflow/axonflow-go/issues
- **Email**: dev@getaxonflow.com

## License

MIT
