# LDAP Operations Testing Framework - Development Prompt

## Project Overview

I need to build an automated testing application in Go that validates LDAP operations against a PingDS LDAP server. The application should intelligently test all core LDAP operations on real directory objects.

## Requirements

### Core Functionality

1. **Test all 9 LDAP operations:**
   - Bind (authentication)
   - Search (query entries)
   - Compare (attribute value testing)
   - Add (create entries)
   - Delete (remove entries)
   - Modify (update attributes)
   - Modify DN (rename/move entries)
   - Unbind (disconnect)
   - Abandon (cancel operations)

2. **Intelligent Test Design:**
   - Tests should operate on real LDAP objects (not just mock data)
   - Create a test organizational structure (OUs, users, groups)
   - Each operation should test realistic scenarios
   - Include both positive tests (expected to succeed) and negative tests (expected to fail gracefully)
   - Tests should be idempotent where possible

3. **Test Flow:**
   - Setup phase: Create test organizational structure and test entries
   - Execution phase: Run all operation tests
   - Validation phase: Verify results of each operation
   - Cleanup phase: **ONLY remove test data if `--cleanup` flag is explicitly specified** (default: preserve data)

## Logging System (CRITICAL REQUIREMENT)

### Multi-Level Logging

- **ERROR:** Critical failures that prevent test execution
- **WARN:** Non-fatal issues or unexpected behavior
- **INFO:** High-level test progress and results (default level)
- **DEBUG:** Detailed operation information
- **TRACE:** Complete LDAP protocol details (filters, DNs, attributes, response codes)

### Logging Features

- Structured logging with timestamps, log levels, and context
- Log to both console (with color coding) and file simultaneously
- Configurable log levels via CLI flag (e.g., `--log-level=debug`)
- Verbose mode flag (`--verbose` or `-v`) sets log level to TRACE
- Each LDAP operation should log:
  - The exact operation being performed
  - Full DN of target entry
  - All attributes being sent/modified
  - LDAP filters used in searches
  - Response codes and messages
  - Duration of operation
  - Any errors or warnings

### Audit Trail

- Every test action should be logged with enough detail to reproduce it manually
- Log file should be named with timestamp (e.g., `ldap-test-2025-11-03-14-30-45.log`)
- Option to specify custom log file path via `--log-file` flag
- Rolling log files option for repeated test runs

### Example Log Output (TRACE level)

```
2025-11-03 14:30:45.123 INFO  [TestRunner] Starting LDAP operations test suite
2025-11-03 14:30:45.124 DEBUG [Connection] Connecting to ldap://ldap.example.com:389
2025-11-03 14:30:45.156 DEBUG [Bind] Attempting bind with DN: cn=admin,dc=example,dc=com
2025-11-03 14:30:45.198 INFO  [Bind] Successfully authenticated
2025-11-03 14:30:45.199 TRACE [Setup] Creating test OU: ou=test-2025-11-03,dc=example,dc=com
2025-11-03 14:30:45.200 TRACE [Add] Operation: Add
2025-11-03 14:30:45.200 TRACE [Add] DN: ou=test-2025-11-03,dc=example,dc=com
2025-11-03 14:30:45.200 TRACE [Add] Attributes: {objectClass: [organizationalUnit], ou: [test-2025-11-03]}
2025-11-03 14:30:45.245 TRACE [Add] Response: Success (code: 0), Duration: 45ms
2025-11-03 14:30:45.246 INFO  [Setup] Test OU created successfully
2025-11-03 14:30:45.247 TRACE [Search] Operation: Search
2025-11-03 14:30:45.247 TRACE [Search] Base DN: ou=test-2025-11-03,dc=example,dc=com
2025-11-03 14:30:45.247 TRACE [Search] Filter: (objectClass=*)
2025-11-03 14:30:45.247 TRACE [Search] Scope: base, Attributes: [*]
2025-11-03 14:30:45.289 TRACE [Search] Found 1 entries, Duration: 42ms
```

## Data Persistence (CRITICAL REQUIREMENT)

### Default Behavior: PRESERVE DATA

- By default, the application **NEVER** deletes test data
- All created entries remain in LDAP after tests complete
- This allows for manual inspection and debugging

### Cleanup Flag

- `--cleanup` flag explicitly enables data deletion
- When specified, cleanup runs in reverse order of creation
- Log each deletion with full DN
- If cleanup fails, log the error but don't fail the test run
- Optional `--cleanup-on-success` to only cleanup if all tests pass

### Data Management

- Use timestamped prefixes for test data (e.g., `test-20251103-143045-`)
- Option to list existing test data: `--list-test-data`
- Option to cleanup old test runs: `--cleanup-older-than=7d`
- Track all created DNs in memory for cleanup operations

## Technical Specifications

### Library

Use `github.com/go-ldap/ldap/v3` (https://github.com/go-ldap/ldap)

### Architecture

- Clean, modular code with separation of concerns
- Configuration via environment variables or config file (YAML/JSON)
- Comprehensive error handling and logging
- Test results should be reported in a clear format (console output + optional JSON/XML report)

### Configuration Parameters

- LDAP server host/port
- Bind DN and credentials
- Base DN for test operations
- TLS/StartTLS options
- Test prefix (to namespace test entries)
- Timeout values
- Log level and log file path

### CLI Flags

```
--config, -c         Config file path (default: ./ldap-test-config.yaml)
--host               LDAP server host
--port               LDAP server port (default: 389)
--bind-dn            Bind DN for authentication
--bind-password      Bind password
--base-dn            Base DN for test operations
--use-tls            Use LDAPS (default: false)
--start-tls          Use StartTLS (default: false)
--log-level          Log level: error|warn|info|debug|trace (default: info)
--log-file           Log file path (default: ./logs/ldap-test-{timestamp}.log)
--verbose, -v        Enable verbose logging (sets log-level to trace)
--cleanup            Delete test data after run (default: false)
--cleanup-on-success Delete test data only if all tests pass
--list-test-data     List existing test data and exit
--cleanup-older-than Cleanup test data older than duration (e.g., 7d, 24h)
--dry-run            Preview operations without executing
--test-suite         Run specific test suite (add|modify|search|all)
--concurrent         Number of concurrent test workers (default: 1)
--report-format      Output format: console|json|xml (default: console)
```

## Test Coverage

Each operation should include multiple test cases:

- **Bind:** Test valid credentials, invalid credentials, anonymous bind
- **Search:** Test various filters, scopes (base, one, sub), attribute selection, paging
- **Compare:** Test matching and non-matching attribute values
- **Add:** Create users, groups, OUs with various attribute combinations
- **Modify:** Add/replace/delete attributes on existing entries
- **Modify DN:** Rename entries, move entries between OUs
- **Delete:** Remove individual entries, test cascade scenarios (only if --cleanup specified)
- **Abandon:** Cancel long-running search operations

## Output Requirements

- Clear pass/fail status for each test
- Detailed error messages on failures
- Statistics summary (total tests, passed, failed, duration)
- Verbose mode for detailed operation logs
- Exit code indicating overall success/failure
- Summary of created objects with DNs (if data preserved)

## Additional Features

- Concurrent test execution option (with appropriate synchronization)
- Dry-run mode to preview what will be tested
- Ability to run specific test suites or individual tests
- Health check/connectivity validation before running tests
- Support for both LDAP and LDAPS connections
- Graceful handling of Ctrl+C (cleanup if flag set, otherwise preserve state)

## Logging Library Recommendation

- Use a structured logging library like `logrus`, `zap`, or `zerolog`
- Support for JSON structured logs as an option
- Thread-safe logging for concurrent operations

## Code Quality

- Well-commented code explaining LDAP operations
- README with setup instructions and usage examples
- Examples of configuration files
- Unit tests for utility functions
- Follow Go best practices and idioms

## Expected Deliverables

1. Main application with CLI interface
2. Configuration file example (ldap-test-config.yaml)
3. README.md with:
   - Installation instructions
   - Configuration guide
   - Usage examples with different flags
   - Log level descriptions
   - Troubleshooting guide
4. Sample output showing test results at different log levels
5. Example log files demonstrating verbose output

## Success Criteria

- A developer should be able to examine the log file and understand exactly what operations were performed on which LDAP entries
- Test data should persist by default for manual verification
- The application should be production-ready for CI/CD integration

## Goal

Please create this application with a focus on observability, debuggability, and safe defaults (preserve data unless explicitly told to cleanup).