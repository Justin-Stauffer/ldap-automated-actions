# LDAP Operations Testing Framework

A comprehensive Go-based testing application that validates all LDAP operations against a PingDS LDAP server (or any LDAP-compliant server). This tool provides intelligent testing with real directory objects, extensive logging, and safe defaults.

## Features

- **Complete LDAP Operation Coverage**: Tests all 9 core LDAP operations
  - Bind (authentication)
  - Search (query entries)
  - Compare (attribute value testing)
  - Add (create entries)
  - Delete (remove entries)
  - Modify (update attributes)
  - Modify DN (rename/move entries)
  - Unbind (disconnect)
  - Abandon (cancel operations)

- **Multi-Level Logging**: ERROR, WARN, INFO, DEBUG, and TRACE levels with detailed operation logging
- **Safe Defaults**: Test data is preserved by default for manual inspection
- **Flexible Configuration**: YAML config files and comprehensive CLI flags
- **Real-World Testing**: Creates actual LDAP entries (OUs, users, groups) for realistic testing
- **Health Checks**: Validates LDAP server connectivity before running tests
- **Comprehensive Reporting**: Detailed test results with pass/fail status and timing

## Installation

### Prerequisites

- Go 1.19 or higher
- Access to an LDAP server (PingDS, OpenLDAP, Active Directory, etc.)
- Valid credentials for LDAP authentication

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd ldap-automated-actions

# Build the application
go build -o ldap-test ./cmd/ldap-test

# Or install directly
go install ./cmd/ldap-test
```

### Quick Start

```bash
# Run with default config file
./ldap-test --config configs/ldap-test-config.yaml

# Run with CLI flags
./ldap-test \
  --host ldap.example.com \
  --port 389 \
  --bind-dn "cn=admin,dc=example,dc=com" \
  --bind-password "password" \
  --base-dn "dc=example,dc=com" \
  --log-level debug
```

## Configuration

### Configuration File

Create a YAML configuration file (example in `configs/ldap-test-config.yaml`):

```yaml
# LDAP Server Settings
host: "ldap.example.com"
port: 389
bind_dn: "cn=admin,dc=example,dc=com"
bind_password: "your_password"
base_dn: "dc=example,dc=com"

# TLS Settings
use_tls: false
start_tls: false
timeout: 30

# Test Settings
test_prefix: "ldap-test"
test_suite: "all"
concurrent: 1
dry_run: false

# Logging
log_level: "info"
log_file: "./logs/ldap-test.log"
verbose: false

# Cleanup
cleanup: false
cleanup_on_success: false
```

### CLI Flags

All configuration options can be overridden via CLI flags:

#### Connection Flags
- `--host` - LDAP server hostname
- `--port` - LDAP server port (default: 389)
- `--bind-dn` - DN for authentication
- `--bind-password` - Password for authentication
- `--base-dn` - Base DN for test operations
- `--use-tls` - Use LDAPS (LDAP over TLS)
- `--start-tls` - Use StartTLS
- `--timeout` - Connection timeout in seconds (default: 30)

#### Test Flags
- `--test-prefix` - Prefix for test entries (default: "ldap-test")
- `--test-suite` - Specific test suite to run: `all`, `bind`, `search`, `add`, `modify`, `compare`, `modifydn`, `delete`, `abandon` (default: "all")
- `--concurrent` - Number of concurrent test workers (default: 1)
- `--dry-run` - Preview operations without executing

#### Logging Flags
- `--log-level` - Log level: `error`, `warn`, `info`, `debug`, `trace` (default: "info")
- `--log-file` - Path to log file
- `--verbose`, `-v` - Enable verbose logging (sets log-level to trace)

#### Cleanup Flags
- `--cleanup` - Delete test data after run (default: false)
- `--cleanup-on-success` - Delete test data only if all tests pass
- `--list-test-data` - List existing test data and exit
- `--cleanup-older-than` - Cleanup data older than duration (e.g., "7d", "24h")

#### Other Flags
- `--report-format` - Output format: `console`, `json`, `xml` (default: "console")
- `--config`, `-c` - Config file path (default: "./configs/ldap-test-config.yaml")
- `--version` - Show version information
- `--help`, `-h` - Show help message

## Usage Examples

### Basic Usage

Run all tests with INFO level logging:
```bash
./ldap-test \
  --host ldap.example.com \
  --bind-dn "cn=admin,dc=example,dc=com" \
  --bind-password "password" \
  --base-dn "dc=example,dc=com"
```

### Verbose Logging

Run with full TRACE level logging for debugging:
```bash
./ldap-test \
  --host ldap.example.com \
  --bind-dn "cn=admin,dc=example,dc=com" \
  --bind-password "password" \
  --base-dn "dc=example,dc=com" \
  --verbose
```

### Run Specific Test Suite

Run only search operation tests:
```bash
./ldap-test \
  --config configs/ldap-test-config.yaml \
  --test-suite search \
  --log-level debug
```

### With Automatic Cleanup

Run tests and cleanup all created data:
```bash
./ldap-test \
  --config configs/ldap-test-config.yaml \
  --cleanup
```

### Cleanup Only on Success

Preserve test data if any test fails (for debugging):
```bash
./ldap-test \
  --config configs/ldap-test-config.yaml \
  --cleanup-on-success
```

### Dry Run Mode

Preview what tests will be executed without making changes:
```bash
./ldap-test \
  --config configs/ldap-test-config.yaml \
  --dry-run \
  --verbose
```

### Using TLS/LDAPS

Connect via LDAPS (TLS):
```bash
./ldap-test \
  --host ldaps.example.com \
  --port 636 \
  --use-tls \
  --bind-dn "cn=admin,dc=example,dc=com" \
  --bind-password "password" \
  --base-dn "dc=example,dc=com"
```

Connect with StartTLS:
```bash
./ldap-test \
  --host ldap.example.com \
  --port 389 \
  --start-tls \
  --bind-dn "cn=admin,dc=example,dc=com" \
  --bind-password "password" \
  --base-dn "dc=example,dc=com"
```

## Log Levels

### ERROR
Critical failures that prevent test execution. Example:
```
2025-11-03 14:30:45.198 ERROR [Connection] Failed to connect to LDAP server
```

### WARN
Non-fatal issues or unexpected behavior. Example:
```
2025-11-03 14:30:45.198 WARN  [Cleanup] Failed to delete entry: cn=test,dc=example,dc=com
```

### INFO (Default)
High-level test progress and results. Example:
```
2025-11-03 14:30:45.198 INFO  [BindTest] PASS: Valid Bind Test
2025-11-03 14:30:45.199 INFO  [AddTest] PASS: Add User Test
```

### DEBUG
Detailed operation information. Example:
```
2025-11-03 14:30:45.198 DEBUG [Connection] Connecting to ldap://ldap.example.com:389
2025-11-03 14:30:45.199 DEBUG [Bind] Attempting bind with DN: cn=admin,dc=example,dc=com
```

### TRACE
Complete LDAP protocol details (filters, DNs, attributes, response codes). Example:
```
2025-11-03 14:30:45.200 TRACE [Add] Operation: Add
2025-11-03 14:30:45.200 TRACE [Add] DN: ou=test-2025-11-03,dc=example,dc=com
2025-11-03 14:30:45.200 TRACE [Add] Attributes: {objectClass: [organizationalUnit], ou: [test-2025-11-03]}
2025-11-03 14:30:45.245 TRACE [Add] Response: Success (code: 0), Duration: 45ms
```

## Test Operations

### Bind Tests
- Valid credentials authentication
- Invalid credentials rejection
- Anonymous bind handling

### Add Tests
- Create organizational units (OUs)
- Create user entries (inetOrgPerson)
- Create group entries (groupOfNames)
- Duplicate entry detection
- Missing required attributes validation

### Search Tests
- Base scope search
- One-level scope search
- Subtree scope search
- Filter-based search
- Attribute selection
- Paged results

### Modify Tests
- Add attribute values
- Replace attribute values
- Delete attribute values
- Multiple modifications in one request
- Non-existent entry handling

### Compare Tests
- Matching attribute values
- Non-matching attribute values
- Non-existent entry handling
- Non-existent attribute handling

### Modify DN Tests
- Rename entries (change RDN)
- Move entries between OUs
- Rename and move simultaneously
- Existing DN conflict detection

### Delete Tests
- Delete leaf entries
- Non-leaf entry protection
- Non-existent entry handling

### Abandon Tests
- Cancel long-running operations

### Unbind Tests
- Clean connection termination

## Output Format

### Console Output (Default)

```
================================================================================
LDAP OPERATIONS TEST SUITE RESULTS
================================================================================
Total Tests:     25
Passed:          24
Failed:          1
Duration:        2.5s
================================================================================

Detailed Results:
--------------------------------------------------------------------------------

Bind Tests:
  ✓ PASS  Valid Bind Test                                      45ms
  ✓ PASS  Invalid Bind Test (Negative)                        32ms
  ✓ PASS  Anonymous Bind Test                                  28ms

Add Tests:
  ✓ PASS  Add OU Test                                          67ms
  ✓ PASS  Add User Test                                        89ms
  ✓ PASS  Add Group Test                                       72ms
  ✓ PASS  Add Duplicate Entry Test (Negative)                 34ms
  ✓ PASS  Add Entry with Missing Required Attributes Test     41ms

... (more tests)

=== Created Test Data Summary ===
Total entries created: 8

OU entries (2):
  - ou=ldap-test-20251103-143045,dc=example,dc=com
  - ou=target-ou,ou=ldap-test-20251103-143045,dc=example,dc=com

User entries (4):
  - cn=testuser,ou=ldap-test-20251103-143045,dc=example,dc=com
  - cn=renamed-user,ou=ldap-test-20251103-143045,dc=example,dc=com
  - cn=move-test-user,ou=target-ou,ou=ldap-test-20251103-143045,dc=example,dc=com
  - cn=renamed-moved-user,ou=target-ou,ou=ldap-test-20251103-143045,dc=example,dc=com

Group entries (2):
  - cn=testgroup,ou=ldap-test-20251103-143045,dc=example,dc=com
  - cn=admins,ou=ldap-test-20251103-143045,dc=example,dc=com

Note: Test data has been preserved. Use --cleanup flag to remove it automatically.

================================================================================
✓ ALL TESTS PASSED
================================================================================
```

## Troubleshooting

### Connection Issues

**Problem**: `Failed to connect to LDAP server`

**Solutions**:
- Verify host and port are correct
- Check firewall rules
- Try `--use-tls` or `--start-tls` if required
- Increase `--timeout` value

### Authentication Issues

**Problem**: `Bind failed: Invalid credentials`

**Solutions**:
- Verify bind DN format (e.g., `cn=admin,dc=example,dc=com`)
- Check password is correct
- Ensure the account has necessary permissions

### Permission Issues

**Problem**: `Insufficient access rights`

**Solutions**:
- Verify bind user has write permissions to base DN
- Check ACLs on LDAP server
- Use an administrator account for testing

### Test Data Not Cleaned Up

**Problem**: Test entries remain after run

**Solutions**:
- This is expected behavior (safe default)
- Use `--cleanup` flag to automatically remove test data
- Use `--cleanup-on-success` to remove only if all tests pass
- Manually delete entries using LDAP admin tools

## CI/CD Integration

### Exit Codes
- `0`: All tests passed
- `1`: One or more tests failed or execution error

### Example CI Pipeline

```yaml
# .gitlab-ci.yml
test-ldap:
  stage: test
  script:
    - ./ldap-test --config ci-config.yaml --cleanup-on-success
  artifacts:
    reports:
      junit: logs/ldap-test-*.log
```

```yaml
# GitHub Actions
- name: Run LDAP Tests
  run: |
    ./ldap-test \
      --host ${{ secrets.LDAP_HOST }} \
      --bind-dn ${{ secrets.LDAP_BIND_DN }} \
      --bind-password ${{ secrets.LDAP_BIND_PASSWORD }} \
      --base-dn ${{ secrets.LDAP_BASE_DN }} \
      --cleanup-on-success
```

## Development

### Project Structure

```
ldap-automated-actions/
├── cmd/
│   └── ldap-test/          # Main application
│       └── main.go
├── internal/
│   ├── config/             # Configuration handling
│   │   └── config.go
│   ├── logger/             # Logging system
│   │   └── logger.go
│   ├── ldap/               # LDAP connection management
│   │   └── connection.go
│   ├── tests/              # Test implementations
│   │   ├── runner.go
│   │   ├── types.go
│   │   ├── bind.go
│   │   ├── add.go
│   │   ├── search.go
│   │   ├── modify.go
│   │   ├── compare.go
│   │   ├── modifydn.go
│   │   ├── delete.go
│   │   └── abandon.go
│   └── tracker/            # Test data tracking
│       └── tracker.go
├── configs/                # Configuration examples
│   └── ldap-test-config.yaml
├── logs/                   # Log files (auto-created)
├── go.mod
├── go.sum
└── README.md
```

### Building

```bash
# Build for current platform
go build -o ldap-test ./cmd/ldap-test

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o ldap-test-linux ./cmd/ldap-test

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o ldap-test.exe ./cmd/ldap-test

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o ldap-test-macos ./cmd/ldap-test
```

### Running Tests

```bash
# Run Go unit tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
```

## Contributing

Contributions are welcome! Please ensure:
- Code follows Go best practices
- All tests pass
- New features include appropriate tests
- Documentation is updated

## License

[Specify your license here]

## Support

For issues, questions, or contributions, please [create an issue](link-to-issues) on the project repository.

## Acknowledgments

- Built with [go-ldap/ldap](https://github.com/go-ldap/ldap) library
- Uses [logrus](https://github.com/sirupsen/logrus) for structured logging
- CLI powered by [pflag](https://github.com/spf13/pflag)
