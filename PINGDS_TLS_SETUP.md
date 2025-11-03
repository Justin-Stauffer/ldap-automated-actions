# PingDS TLS/SSL Configuration Guide

This guide explains how to use the LDAP testing tool with PingDS servers that use custom certificates and PKCS12 trust stores.

## Background

PingDS (formerly OpenDJ) typically uses self-signed certificates or custom PKI, requiring trust store configuration for secure connections. This tool now supports PKCS12 trust stores, matching the PingDS CLI tools behavior.

## Configuration Options

### 1. Trust Store Path
Specifies the path to your PKCS12 trust store file (the keystore file).

**Config File:**
```yaml
trust_store_path: "C:\\path\\to\\opendj\\config\\keystore"
```

**CLI Flag:**
```bash
--trust-store-path "C:\path\to\opendj\config\keystore"
```

### 2. Trust Store Password

You can provide the password in two ways:

#### Option A: Direct Password (Less Secure)
```yaml
trust_store_password: "your_password"
```
```bash
--trust-store-password "your_password"
```

#### Option B: Password File (More Secure)
```yaml
trust_store_password_file: "C:\\path\\to\\opendj\\config\\keystore.pin"
```
```bash
--trust-store-password-file "C:\path\to\opendj\config\keystore.pin"
```

### 3. Skip Certificate Verification (Testing Only)
For development/testing environments, you can skip certificate verification:

```yaml
insecure_skip_verify: true
```
```bash
--insecure-skip-verify
```

⚠️ **WARNING**: Never use `insecure_skip_verify` in production!

## Example Configurations

### Example 1: PingDS with LDAPS (Recommended)

**Config File:**
```yaml
host: "localhost"
port: 1636
bind_dn: "uid=admin,ou=People,dc=example,dc=com"
bind_password: "password"
base_dn: "dc=example,dc=com"

# Use LDAPS with custom certificate
use_tls: true
trust_store_path: "C:\\opendj\\config\\keystore"
trust_store_password_file: "C:\\opendj\\config\\keystore.pin"
```

**CLI:**
```bash
./ldap-test \
  --host localhost \
  --port 1636 \
  --use-tls \
  --trust-store-path "C:\opendj\config\keystore" \
  --trust-store-password-file "C:\opendj\config\keystore.pin" \
  --bind-dn "uid=admin,ou=People,dc=example,dc=com" \
  --bind-password "password" \
  --base-dn "dc=example,dc=com"
```

### Example 2: StartTLS with Trust Store

**Config File:**
```yaml
host: "ldap.example.com"
port: 389
bind_dn: "cn=admin,dc=example,dc=com"
bind_password: "password"
base_dn: "dc=example,dc=com"

# Start with plain LDAP, upgrade to TLS
start_tls: true
trust_store_path: "/opt/opendj/config/keystore"
trust_store_password_file: "/opt/opendj/config/keystore.pin"
```

**CLI:**
```bash
./ldap-test \
  --host ldap.example.com \
  --port 389 \
  --start-tls \
  --trust-store-path "/opt/opendj/config/keystore" \
  --trust-store-password-file "/opt/opendj/config/keystore.pin" \
  --bind-dn "cn=admin,dc=example,dc=com" \
  --bind-password "password" \
  --base-dn "dc=example,dc=com"
```

### Example 3: Testing Environment (Skip Verification)

**For testing only** - skips certificate validation:

```bash
./ldap-test \
  --host localhost \
  --port 1636 \
  --use-tls \
  --insecure-skip-verify \
  --bind-dn "cn=admin,dc=example,dc=com" \
  --bind-password "password" \
  --base-dn "dc=example,dc=com"
```

## Comparison with PingDS CLI Tools

### PingDS ldapsearch Command:
```bash
ldapsearch.bat `
  --hostname localhost `
  --port 1636 `
  --useSsl `
  --usePkcs12TrustStore C:\path\to\opendj\config\keystore `
  --trustStorePassword:file C:\path\to\opendj\config\keystore.pin `
  --bindDN uid=bjensen,ou=People,dc=example,dc=com `
  --bindPassword hifalutin `
  --baseDn dc=example,dc=com `
  "(cn=Babs Jensen)" `
  cn mail street l
```

### Equivalent LDAP Test Tool Command:
```bash
./ldap-test.exe `
  --host localhost `
  --port 1636 `
  --use-tls `
  --trust-store-path "C:\path\to\opendj\config\keystore" `
  --trust-store-password-file "C:\path\to\opendj\config\keystore.pin" `
  --bind-dn "uid=bjensen,ou=People,dc=example,dc=com" `
  --bind-password hifalutin `
  --base-dn "dc=example,dc=com" `
  --test-suite search `
  --verbose
```

## Logging Output

When using trust stores, you'll see detailed logging at DEBUG and TRACE levels:

```
2025-11-03 14:30:45.123 DEBUG [TLS] Loading PKCS12 trust store, path=C:\opendj\config\keystore
2025-11-03 14:30:45.124 DEBUG [TLS] Reading trust store password from file, file=C:\opendj\config\keystore.pin
2025-11-03 14:30:45.156 TRACE [TLS] Added certificate to pool, subject=CN=localhost
2025-11-03 14:30:45.157 INFO  [TLS] Loaded trust store, certificates=3
2025-11-03 14:30:45.198 INFO  [Connection] Successfully connected to LDAP server
```

If there are certificate issues:
```
2025-11-03 14:30:45.456 ERROR [Connection] Failed to connect to LDAP server, error=x509: certificate signed by unknown authority
```

## Troubleshooting

### Issue: "failed to decode PKCS12 trust store"
**Solution**: Check that:
- The trust store path is correct
- The password is correct
- The file is a valid PKCS12 keystore

### Issue: "x509: certificate signed by unknown authority"
**Solutions**:
1. Verify you're using the correct trust store file
2. Check that the certificate is in the trust store
3. For testing only: use `--insecure-skip-verify`

### Issue: "failed to read trust store password file"
**Solution**:
- Verify the password file path is correct
- Check file permissions
- Ensure the file contains only the password (no extra whitespace)

### Issue: "No certificates found in trust store"
**Solution**:
- Verify the PKCS12 file contains certificates
- Try using the PingDS tools with the same keystore to verify it works

## Security Best Practices

1. ✅ **Use password files** instead of embedding passwords in config or CLI
2. ✅ **Use LDAPS (port 636)** or **StartTLS (port 389)** for production
3. ✅ **Keep trust stores protected** with appropriate file permissions
4. ✅ **Rotate certificates** according to your security policy
5. ❌ **Never use `insecure_skip_verify`** in production environments
6. ❌ **Don't commit** trust store files or passwords to version control

## Additional Resources

- PingDS Documentation: https://docs.pingidentity.com/
- PKCS12 Format: https://en.wikipedia.org/wiki/PKCS_12
- Go PKCS12 Library: https://pkg.go.dev/software.sslmate.com/src/go-pkcs12

## Testing Your Configuration

To test your trust store configuration without running full tests:

```bash
# Dry run mode
./ldap-test --config your-config.yaml --dry-run --verbose

# Just connection and health check (will fail after bind, but verifies TLS works)
./ldap-test --config your-config.yaml --test-suite bind --verbose
```

Look for these success indicators in the logs:
- `Loaded trust store, certificates=X` (where X > 0)
- `Successfully connected to LDAP server`
- `Successfully authenticated`

---

For more information, see the main README.md file.
