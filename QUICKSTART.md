# Quick Start Guide

## Prerequisites

1. Go 1.19+ installed
2. Access to an LDAP server
3. Valid LDAP credentials with write permissions

## Build

```bash
go build -o ldap-test ./cmd/ldap-test
```

## Basic Usage

### 1. Configure Your LDAP Server

Edit `configs/ldap-test-config.yaml`:

```yaml
host: "your-ldap-server.com"
port: 389
bind_dn: "cn=admin,dc=example,dc=com"
bind_password: "your_password"
base_dn: "dc=example,dc=com"
log_level: "info"
```

### 2. Run All Tests

```bash
./ldap-test --config configs/ldap-test-config.yaml
```

### 3. Run with Verbose Logging

```bash
./ldap-test --config configs/ldap-test-config.yaml --verbose
```

### 4. Run Specific Test Suite

```bash
./ldap-test --config configs/ldap-test-config.yaml --test-suite bind
./ldap-test --config configs/ldap-test-config.yaml --test-suite search
./ldap-test --config configs/ldap-test-config.yaml --test-suite add
```

### 5. Run with Cleanup

```bash
# Always cleanup
./ldap-test --config configs/ldap-test-config.yaml --cleanup

# Cleanup only if all tests pass
./ldap-test --config configs/ldap-test-config.yaml --cleanup-on-success
```

## Using CLI Flags Only (No Config File)

```bash
./ldap-test \
  --host ldap.example.com \
  --port 389 \
  --bind-dn "cn=admin,dc=example,dc=com" \
  --bind-password "password" \
  --base-dn "dc=example,dc=com" \
  --log-level debug
```

## Common Scenarios

### Development/Testing
```bash
# Run with verbose logging, preserve data for inspection
./ldap-test --config configs/ldap-test-config.yaml --verbose
```

### CI/CD Pipeline
```bash
# Run all tests, cleanup on success, exit with appropriate code
./ldap-test --config configs/ldap-test-config.yaml --cleanup-on-success
```

### Troubleshooting Connection Issues
```bash
# Preview without making changes
./ldap-test --config configs/ldap-test-config.yaml --dry-run --verbose
```

### Testing Specific Operations
```bash
# Test only search functionality with debug output
./ldap-test --config configs/ldap-test-config.yaml --test-suite search --log-level debug
```

## Expected Output

```
================================================================================
LDAP OPERATIONS TEST SUITE RESULTS
================================================================================
Total Tests:     25
Passed:          25
Failed:          0
Duration:        2.345s
================================================================================

Detailed Results:
--------------------------------------------------------------------------------

Bind Tests:
  ✓ PASS  Valid Bind Test                                      45ms
  ✓ PASS  Invalid Bind Test (Negative)                         32ms
  ✓ PASS  Anonymous Bind Test                                  28ms

... (more test results)

=== Created Test Data Summary ===
Total entries created: 8

OU entries (2):
  - ou=ldap-test-20251103-143045,dc=example,dc=com
  - ou=target-ou,ou=ldap-test-20251103-143045,dc=example,dc=com

User entries (4):
  - cn=testuser,ou=ldap-test-20251103-143045,dc=example,dc=com
  ...

Note: Test data has been preserved. Use --cleanup flag to remove it automatically.

================================================================================
✓ ALL TESTS PASSED
================================================================================
```

## Log Files

Logs are saved to `./logs/ldap-test-TIMESTAMP.log` with all LDAP operations recorded.

View logs:
```bash
cat logs/ldap-test-*.log
# or
tail -f logs/ldap-test-*.log  # Watch in real-time
```

## Need Help?

```bash
./ldap-test --help
./ldap-test --version
```

See README.md for comprehensive documentation.
