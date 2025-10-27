package main

import (
	"fmt"
	"log"
	"os"

	"github.com/getaxonflow/axonflow-go"
)

func main() {
	// Load configuration from environment variables
	agentURL := getEnv("AXONFLOW_AGENT_URL", "https://staging-eu.getaxonflow.com")
	clientID := getEnv("AXONFLOW_CLIENT_ID", "")
	clientSecret := getEnv("AXONFLOW_CLIENT_SECRET", "")

	if clientID == "" || clientSecret == "" {
		log.Fatal("AXONFLOW_CLIENT_ID and AXONFLOW_CLIENT_SECRET must be set")
	}

	// Create client with simple initialization
	fmt.Println("Initializing AxonFlow client...")
	client := axonflow.NewClientSimple(agentURL, clientID, clientSecret)

	// Perform health check
	fmt.Println("\nChecking AxonFlow Agent health...")
	if err := client.HealthCheck(); err != nil {
		log.Printf("Warning: Health check failed: %v", err)
	} else {
		fmt.Println("✓ AxonFlow Agent is healthy")
	}

	// Execute a simple query
	fmt.Println("\nExecuting governed query...")
	resp, err := client.ExecuteQuery(
		"demo-user-token",
		"What is the capital of France?",
		"chat",
		map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  100,
		},
	)

	if err != nil {
		log.Fatalf("Query execution failed: %v", err)
	}

	// Check if request was blocked
	if resp.Blocked {
		fmt.Printf("❌ Request blocked by governance policy\n")
		fmt.Printf("   Reason: %s\n", resp.BlockReason)
		fmt.Printf("   Policies evaluated: %v\n", resp.PolicyInfo.PoliciesEvaluated)
		return
	}

	// Check if request succeeded
	if !resp.Success {
		fmt.Printf("❌ Query failed: %s\n", resp.Error)
		return
	}

	// Display result
	fmt.Println("✓ Query executed successfully")
	fmt.Printf("Result: %s\n", resp.Data)

	// Display governance metadata
	fmt.Println("\nGovernance Metadata:")
	fmt.Printf("  Request ID: %s\n", resp.RequestID)
	fmt.Printf("  Policies Evaluated: %v\n", resp.PolicyInfo.PoliciesEvaluated)
	fmt.Printf("  Processing Time: %dms\n", resp.Metadata.ProcessingTime)

	// Test with sensitive data (should be redacted)
	fmt.Println("\n" + "="*60)
	fmt.Println("Testing PII detection and redaction...")
	fmt.Println("="*60)

	resp2, err := client.ExecuteQuery(
		"demo-user-token",
		"My email is john.doe@example.com and my SSN is 123-45-6789",
		"chat",
		map[string]interface{}{},
	)

	if err != nil {
		log.Printf("Warning: PII test query failed: %v", err)
		return
	}

	if resp2.Blocked {
		fmt.Printf("✓ PII detected and request blocked\n")
		fmt.Printf("  Reason: %s\n", resp2.BlockReason)
	} else {
		fmt.Printf("✓ PII handled: %s\n", resp2.Data)
	}
}

// getEnv retrieves environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
