package ldap

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"ldap-automated-actions/internal/config"
	"ldap-automated-actions/internal/logger"

	"github.com/go-ldap/ldap/v3"
	"software.sslmate.com/src/go-pkcs12"
)

// Connection represents an LDAP connection wrapper
type Connection struct {
	conn   *ldap.Conn
	config *config.Config
}

// buildTLSConfig creates a TLS configuration based on the provided config
func buildTLSConfig(cfg *config.Config) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		ServerName:         cfg.Host,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}

	// If trust store is specified, load it
	if cfg.TrustStorePath != "" {
		logger.Debug("TLS", "Loading PKCS12 trust store", "path", cfg.TrustStorePath)

		// Read trust store password
		password := cfg.TrustStorePassword
		if cfg.TrustStorePasswordFile != "" {
			logger.Debug("TLS", "Reading trust store password from file", "file", cfg.TrustStorePasswordFile)
			passwordBytes, err := os.ReadFile(cfg.TrustStorePasswordFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read trust store password file: %w", err)
			}
			password = strings.TrimSpace(string(passwordBytes))
		}

		// Read PKCS12 file
		p12Data, err := os.ReadFile(cfg.TrustStorePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read trust store file: %w", err)
		}

		// Decode PKCS12
		blocks, err := pkcs12.ToPEM(p12Data, password)
		if err != nil {
			return nil, fmt.Errorf("failed to decode PKCS12 trust store: %w", err)
		}

		// Create certificate pool
		certPool := x509.NewCertPool()
		certsAdded := 0

		for _, block := range blocks {
			if block.Type == "CERTIFICATE" {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					logger.Warn("TLS", "Failed to parse certificate in trust store", "error", err)
					continue
				}
				certPool.AddCert(cert)
				certsAdded++
				logger.Trace("TLS", "Added certificate to pool", "subject", cert.Subject.CommonName)
			}
		}

		if certsAdded > 0 {
			tlsConfig.RootCAs = certPool
			logger.Info("TLS", "Loaded trust store", "certificates", certsAdded)
		} else {
			logger.Warn("TLS", "No certificates found in trust store")
		}
	}

	if cfg.InsecureSkipVerify {
		logger.Warn("TLS", "Certificate verification is DISABLED - not recommended for production")
	}

	return tlsConfig, nil
}

// NewConnection creates a new LDAP connection
func NewConnection(cfg *config.Config) (*Connection, error) {
	logger.Debug("Connection", "Attempting to connect to LDAP server", "address", cfg.GetAddress())

	var conn *ldap.Conn
	var err error

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	if cfg.UseTLS {
		// Use LDAPS (LDAP over TLS)
		tlsConfig, err := buildTLSConfig(cfg)
		if err != nil {
			logger.Error("Connection", "Failed to build TLS configuration", "error", err)
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}

		conn, err = ldap.DialTLS("tcp", address, tlsConfig)
	} else {
		// Use plain LDAP
		conn, err = ldap.Dial("tcp", address)
	}

	if err != nil {
		logger.Error("Connection", "Failed to connect to LDAP server", "error", err, "address", address)
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Set timeout
	if cfg.Timeout > 0 {
		conn.SetTimeout(time.Duration(cfg.Timeout) * time.Second)
	}

	// Use StartTLS if configured
	if cfg.StartTLS && !cfg.UseTLS {
		tlsConfig, err := buildTLSConfig(cfg)
		if err != nil {
			conn.Close()
			logger.Error("Connection", "Failed to build TLS configuration for StartTLS", "error", err)
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}

		if err := conn.StartTLS(tlsConfig); err != nil {
			conn.Close()
			logger.Error("Connection", "Failed to start TLS", "error", err)
			return nil, fmt.Errorf("failed to start TLS: %w", err)
		}
		logger.Debug("Connection", "StartTLS successful")
	}

	logger.Info("Connection", "Successfully connected to LDAP server", "address", cfg.GetAddress())

	return &Connection{
		conn:   conn,
		config: cfg,
	}, nil
}

// Bind authenticates with the LDAP server
func (c *Connection) Bind() error {
	logger.Debug("Bind", "Attempting bind", "dn", c.config.BindDN)

	start := time.Now()
	err := c.conn.Bind(c.config.BindDN, c.config.BindPassword)
	duration := time.Since(start)

	if err != nil {
		logger.LogLDAPResult("Bind", "Bind", false, -1, err.Error(), duration)
		return fmt.Errorf("bind failed: %w", err)
	}

	logger.LogLDAPResult("Bind", "Bind", true, 0, "Success", duration)
	logger.Info("Bind", "Successfully authenticated", "dn", c.config.BindDN)

	return nil
}

// Close closes the LDAP connection
func (c *Connection) Close() {
	if c.conn != nil {
		logger.Debug("Connection", "Closing LDAP connection")
		c.conn.Close()
	}
}

// Unbind sends an unbind request and closes the connection
func (c *Connection) Unbind() error {
	if c.conn != nil {
		logger.Debug("Connection", "Sending unbind request")
		start := time.Now()
		err := c.conn.Unbind()
		duration := time.Since(start)

		if err != nil {
			logger.LogLDAPResult("Unbind", "Unbind", false, -1, err.Error(), duration)
			return fmt.Errorf("unbind failed: %w", err)
		}

		logger.LogLDAPResult("Unbind", "Unbind", true, 0, "Success", duration)
		return nil
	}
	return nil
}

// HealthCheck performs a basic health check on the LDAP connection
func (c *Connection) HealthCheck() error {
	logger.Info("HealthCheck", "Performing LDAP connection health check")

	// Try to search for the root DSE
	searchRequest := ldap.NewSearchRequest(
		"", // Base DN (empty for root DSE)
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=*)",
		[]string{"namingContexts", "supportedLDAPVersion"},
		nil,
	)

	start := time.Now()
	result, err := c.conn.Search(searchRequest)
	duration := time.Since(start)

	if err != nil {
		logger.LogLDAPResult("HealthCheck", "Search", false, -1, err.Error(), duration)
		return fmt.Errorf("health check failed: %w", err)
	}

	logger.LogLDAPResult("HealthCheck", "Search", true, 0, "Success", duration)

	if len(result.Entries) > 0 {
		entry := result.Entries[0]
		logger.Info("HealthCheck", "LDAP server is healthy", "entries", len(result.Entries))

		// Log server capabilities
		if namingContexts := entry.GetAttributeValues("namingContexts"); len(namingContexts) > 0 {
			logger.Debug("HealthCheck", "Naming contexts available", "contexts", namingContexts)
		}
		if versions := entry.GetAttributeValues("supportedLDAPVersion"); len(versions) > 0 {
			logger.Debug("HealthCheck", "Supported LDAP versions", "versions", versions)
		}
	}

	return nil
}

// GetConnection returns the underlying LDAP connection
func (c *Connection) GetConnection() *ldap.Conn {
	return c.conn
}

// GetConfig returns the configuration
func (c *Connection) GetConfig() *config.Config {
	return c.config
}
