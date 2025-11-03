package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the LDAP test application
type Config struct {
	// LDAP Connection Settings
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	BindDN       string `yaml:"bind_dn"`
	BindPassword string `yaml:"bind_password"`
	BaseDN       string `yaml:"base_dn"`
	UseTLS       bool   `yaml:"use_tls"`
	StartTLS     bool   `yaml:"start_tls"`
	Timeout      int    `yaml:"timeout"` // seconds

	// TLS/Certificate Settings
	TrustStorePath         string `yaml:"trust_store_path"`          // Path to PKCS12 trust store file
	TrustStorePassword     string `yaml:"trust_store_password"`      // Trust store password
	TrustStorePasswordFile string `yaml:"trust_store_password_file"` // File containing trust store password
	TLSCertFile            string `yaml:"tls_cert_file"`             // Path to PEM certificate file (alternative to PKCS12)
	TLSCAFile              string `yaml:"tls_ca_file"`               // Path to PEM CA certificate file
	InsecureSkipVerify     bool   `yaml:"insecure_skip_verify"`      // Skip certificate verification (not recommended for production)

	// Test Settings
	TestPrefix string `yaml:"test_prefix"`
	Concurrent int    `yaml:"concurrent"`
	TestSuite  string `yaml:"test_suite"`
	DryRun     bool   `yaml:"dry_run"`

	// Logging Settings
	LogLevel  string `yaml:"log_level"`
	LogFile   string `yaml:"log_file"`
	Verbose   bool   `yaml:"verbose"`

	// Cleanup Settings
	Cleanup          bool   `yaml:"cleanup"`
	CleanupOnSuccess bool   `yaml:"cleanup_on_success"`
	ListTestData     bool   `yaml:"list_test_data"`
	CleanupOlderThan string `yaml:"cleanup_older_than"`

	// Report Settings
	ReportFormat string `yaml:"report_format"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Host:         "localhost",
		Port:         389,
		UseTLS:       false,
		StartTLS:     false,
		Timeout:      30,
		TestPrefix:   "ldap-test",
		Concurrent:   1,
		TestSuite:    "all",
		LogLevel:     "info",
		LogFile:      fmt.Sprintf("./logs/ldap-test-%s.log", time.Now().Format("2006-01-02-15-04-05")),
		Verbose:      false,
		Cleanup:      false,
		ReportFormat: "console",
	}
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist, return default config (not an error)
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if c.BindDN == "" {
		return fmt.Errorf("bind DN is required")
	}
	if c.BindPassword == "" {
		return fmt.Errorf("bind password is required")
	}
	if c.BaseDN == "" {
		return fmt.Errorf("base DN is required")
	}
	if c.UseTLS && c.StartTLS {
		return fmt.Errorf("cannot use both TLS and StartTLS")
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"error": true,
		"warn":  true,
		"info":  true,
		"debug": true,
		"trace": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s (must be error, warn, info, debug, or trace)", c.LogLevel)
	}

	// Validate test suite
	validTestSuites := map[string]bool{
		"all":      true,
		"bind":     true,
		"search":   true,
		"add":      true,
		"modify":   true,
		"compare":  true,
		"modifydn": true,
		"delete":   true,
		"abandon":  true,
	}
	if !validTestSuites[c.TestSuite] {
		return fmt.Errorf("invalid test suite: %s", c.TestSuite)
	}

	// Validate report format
	validReportFormats := map[string]bool{
		"console": true,
		"json":    true,
		"xml":     true,
	}
	if !validReportFormats[c.ReportFormat] {
		return fmt.Errorf("invalid report format: %s (must be console, json, or xml)", c.ReportFormat)
	}

	return nil
}

// GetAddress returns the full LDAP server address
func (c *Config) GetAddress() string {
	protocol := "ldap"
	if c.UseTLS {
		protocol = "ldaps"
	}
	return fmt.Sprintf("%s://%s:%d", protocol, c.Host, c.Port)
}
