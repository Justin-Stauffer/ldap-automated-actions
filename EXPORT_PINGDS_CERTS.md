# Exporting PEM Certificates from PingDS/OpenDJ

This guide explains how to export certificates from PingDS PKCS12 keystores to PEM format for use with the LDAP testing tool.

## Why PEM Format?

PingDS/OpenDJ keystores use proprietary Oracle/Sun attributes that aren't compatible with standard Go PKCS12 libraries. Converting to PEM format resolves this issue.

## Method 1: Using OpenSSL (Recommended)

### Export All Certificates to PEM

```bash
# Windows (PowerShell)
openssl pkcs12 -in "C:\path\to\opendj\config\keystore" `
  -out "C:\path\to\opendj\config\certs.pem" `
  -nokeys `
  -passin file:"C:\path\to\opendj\config\keystore.pin"

# Linux/Unix
openssl pkcs12 -in /opt/opendj/config/keystore \
  -out /opt/opendj/config/certs.pem \
  -nokeys \
  -passin file:/opt/opendj/config/keystore.pin
```

### Export Specific Certificate

```bash
# Export by alias (if you know the alias name)
openssl pkcs12 -in keystore \
  -out server-cert.pem \
  -nokeys \
  -name "server-cert" \
  -passin file:keystore.pin
```

### Export CA Certificate Separately

If you need to export just the CA certificate:

```bash
openssl pkcs12 -in keystore \
  -out ca-cert.pem \
  -nokeys \
  -cacerts \
  -passin file:keystore.pin
```

## Method 2: Using Java Keytool

### List Certificates in Keystore

First, see what's in your keystore:

```bash
keytool -list -v -keystore keystore -storepass:file keystore.pin
```

### Export Individual Certificate

```bash
# Export to DER format first
keytool -exportcert \
  -alias "server-cert" \
  -keystore keystore \
  -file server-cert.der \
  -storepass:file keystore.pin

# Convert DER to PEM
openssl x509 -inform DER -in server-cert.der -out server-cert.pem
```

### Export in PEM Format Directly (Java 9+)

```bash
keytool -exportcert \
  -alias "server-cert" \
  -keystore keystore \
  -file server-cert.pem \
  -rfc \
  -storepass:file keystore.pin
```

## Method 3: Quick PowerShell Script

Save this as `export-certs.ps1`:

```powershell
param(
    [Parameter(Mandatory=$true)]
    [string]$KeystorePath,

    [Parameter(Mandatory=$true)]
    [string]$PasswordFile,

    [Parameter(Mandatory=$true)]
    [string]$OutputFile
)

$password = Get-Content $PasswordFile
openssl pkcs12 -in $KeystorePath -out $OutputFile -nokeys -passin pass:$password

if ($LASTEXITCODE -eq 0) {
    Write-Host "Certificates exported successfully to $OutputFile"
} else {
    Write-Error "Failed to export certificates"
}
```

Usage:
```powershell
.\export-certs.ps1 `
  -KeystorePath "C:\opendj\config\keystore" `
  -PasswordFile "C:\opendj\config\keystore.pin" `
  -OutputFile "C:\opendj\config\certs.pem"
```

## Using Exported Certificates with LDAP Test Tool

### Config File Method

```yaml
host: "localhost"
port: 1636
bind_dn: "cn=admin,dc=example,dc=com"
bind_password: "password"
base_dn: "dc=example,dc=com"

use_tls: true
tls_cert_file: "C:\\opendj\\config\\certs.pem"
# OR separate CA and server certs:
# tls_ca_file: "C:\\opendj\\config\\ca-cert.pem"
# tls_cert_file: "C:\\opendj\\config\\server-cert.pem"
```

### CLI Method

```bash
./ldap-test.exe \
  --host localhost \
  --port 1636 \
  --use-tls \
  --tls-cert-file "C:\opendj\config\certs.pem" \
  --bind-dn "cn=admin,dc=example,dc=com" \
  --bind-password "password" \
  --base-dn "dc=example,dc=com"
```

## Verify PEM File Contents

Check that your PEM file contains valid certificates:

```bash
# View certificate details
openssl x509 -in certs.pem -text -noout

# Check if file contains multiple certificates
grep -c "BEGIN CERTIFICATE" certs.pem

# Windows PowerShell equivalent
(Get-Content certs.pem | Select-String "BEGIN CERTIFICATE").Count
```

## Expected PEM Format

A valid PEM certificate file looks like this:

```
-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKJ5Cx9m8x7+MA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
... (more base64 encoded data) ...
-----END CERTIFICATE-----
```

Multiple certificates can be concatenated in the same file.

## Troubleshooting

### Error: "unable to load PKCS12 file"
- Check the keystore path is correct
- Verify the password is correct
- Ensure OpenSSL is installed and in PATH

### Error: "No certificates found"
- The keystore might only contain private keys
- Try the `-cacerts` flag to export CA certificates
- Use `keytool -list` to see what's actually in the keystore

### Error: "wrong tag" or "parse error"
- The keystore file might be corrupted
- Try exporting with Java keytool instead
- Check file permissions

## Common PingDS Certificate Locations

```
Windows:
C:\Program Files\Ping Identity\OpenDJ\config\keystore
C:\Program Files\Ping Identity\OpenDJ\config\keystore.pin

Linux:
/opt/opendj/config/keystore
/opt/opendj/config/keystore.pin
/usr/local/opendj/config/keystore
```

## Quick Test After Export

Test your exported certificates work:

```bash
# Test connection with OpenSSL
openssl s_client -connect localhost:1636 -CAfile certs.pem

# Test with ldap-test tool
./ldap-test \
  --host localhost \
  --port 1636 \
  --use-tls \
  --tls-cert-file certs.pem \
  --bind-dn "cn=admin,dc=example,dc=com" \
  --bind-password "password" \
  --base-dn "dc=example,dc=com" \
  --test-suite bind \
  --verbose
```

Look for these in the logs:
```
DEBUG [TLS] Loading PEM certificate
TRACE [TLS] Added certificate to pool
INFO  [TLS] Loaded PEM certificates, files=1
INFO  [Connection] Successfully connected to LDAP server
```

---

For more information, see PINGDS_TLS_SETUP.md
