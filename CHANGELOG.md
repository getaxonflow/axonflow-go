# Changelog

All notable changes to the AxonFlow Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.8.0] - 2025-12-28

### Added

- **HITL Support**: `ActionRequireApproval` for human oversight policies
  - Use with `CreateStaticPolicy()` to trigger approval workflows
  - Enterprise: Full HITL queue integration
  - Community: Auto-approves immediately

---

## [1.7.0] - 2025-12-28

### Added

- **Code Governance Support** - `CodeArtifact` type for detecting and auditing LLM-generated code (#16)
  - `CodeArtifact` struct in `PolicyEvaluationInfo` with fields:
    - `IsCodeOutput` - Whether response contains code
    - `Language` - Detected programming language (14 supported)
    - `CodeType` - Code category (function, class, script, config, snippet, module)
    - `SizeBytes` - Size of detected code in bytes
    - `LineCount` - Number of lines of code
    - `SecretsDetected` - Count of potential secrets found
    - `UnsafePatterns` - Count of unsafe code patterns
  - Automatic code detection in LLM responses
  - Supports Python, Go, TypeScript, JavaScript, Java, SQL, Ruby, Rust, C/C++, Bash, YAML, JSON, Dockerfile, Terraform

## [1.6.0] - 2025-12-25

### Added

- **Policy CRUD Methods**: Full policy management support for Unified Policy Architecture v2.0.0
  - `ListStaticPolicies()` - List policies with filtering
  - `GetStaticPolicy()` - Get single policy by ID
  - `CreateStaticPolicy()` - Create custom policy
  - `UpdateStaticPolicy()` - Update existing policy
  - `DeleteStaticPolicy()` - Delete policy
  - `ToggleStaticPolicy()` - Enable/disable policy
  - `GetEffectiveStaticPolicies()` - Get merged hierarchy
  - `TestPattern()` - Test regex pattern

- **Policy Override Methods** (Enterprise)
- **Dynamic Policy Methods**
- **New Types**: `StaticPolicy`, `DynamicPolicy`, `PolicyOverride`

## [1.5.1] - 2025-12-23

### Added

- **MAP Timeout Configuration** - New `MapTimeout` config option (default: 120s) for Multi-Agent Planning operations
  - MAP operations involve multiple LLM calls and can take 30-60+ seconds
  - Separate `mapHttpClient` with longer timeout
  - `GeneratePlan()` and `ExecutePlan()` now use the longer MAP timeout

## [1.5.0] - 2025-12-19

### Added

- **LLM Interceptors** - Transparent governance for LLM API calls (#8)
  - `WrapOpenAIClient()` for OpenAI API interception
  - `WrapAnthropicClient()` for Anthropic API interception
  - `WrapGeminiModel()` for Google Generative AI interception
  - Policy enforcement and audit logging for all providers
- Full feature parity with other SDKs for LLM interceptors

## [1.4.1] - 2025-12-15

### Added

- **Contract Testing Suite** - Validates SDK models against real API responses (#7)
  - JSON fixtures for all response types
  - Integration test workflow with GitHub Actions
- Unit tests for `parseTimeWithFallback` helper (#5)

### Fixed

- Datetime parsing with nanosecond precision (#4)
- `GeneratePlan()` and `ExecutePlan()` authentication with explicit user token (#4)

## [1.4.0] - 2025-12-10

### Changed

- Prepare for repository rename to `axonflow-sdk-go`
- Updated module path and documentation

## [1.3.0] - 2025-12-08

### Added

- **Gateway Mode API** - Support for direct LLM calls with policy enforcement (#1)
  - `GetPolicyApprovedContext()` for pre-checks
  - `AuditLLMCall()` for compliance logging
- Self-hosted mode for localhost deployments
  - Skip auth headers for localhost endpoints
  - License key optional for self-hosted
- User token parameter to `QueryConnector()` method

### Fixed

- Formatting in connectors example
- Printf format mismatch in basic example
- Nested error handling in SDK
- `PolicyEvaluationInfo.ProcessingTime` type mismatch

## [1.2.0] - 2025-12-04

### Added

- License-based authentication as primary method
- License key authentication support

## [1.1.0] - 2025-11-27

### Added

- License key authentication support
- Comprehensive examples with license key authentication

## [1.0.0] - 2025-10-27

### Added

- Initial release of AxonFlow Go SDK
- Core client with `ExecuteQuery()` for governed AI calls
- Policy enforcement with `PolicyViolationError`
- Multi-agent planning with `GeneratePlan()` and `ExecutePlan()`
- MCP connector operations (`ListConnectors`, `InstallConnector`, `QueryConnector`)
- Comprehensive type definitions
- Retry logic with exponential backoff
- Response caching with TTL
