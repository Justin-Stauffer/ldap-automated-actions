package main

import (
	"fmt"
	"os"

	"ldap-automated-actions/internal/config"
	"ldap-automated-actions/internal/logger"
	"ldap-automated-actions/internal/tests"

	"github.com/spf13/pflag"
)

const version = "1.0.0"

func main() {
	// Define CLI flags
	configFile := pflag.StringP("config", "c", "./configs/ldap-test-config.yaml", "Config file path")
	host := pflag.String("host", "", "LDAP server host")
	port := pflag.Int("port", 389, "LDAP server port")
	bindDN := pflag.String("bind-dn", "", "Bind DN for authentication")
	bindPassword := pflag.String("bind-password", "", "Bind password")
	baseDN := pflag.String("base-dn", "", "Base DN for test operations")
	useTLS := pflag.Bool("use-tls", false, "Use LDAPS (LDAP over TLS)")
	startTLS := pflag.Bool("start-tls", false, "Use StartTLS")
	timeout := pflag.Int("timeout", 30, "Connection timeout in seconds")

	testPrefix := pflag.String("test-prefix", "ldap-test", "Prefix for test entries")
	testSuite := pflag.String("test-suite", "all", "Test suite to run: all|bind|search|add|modify|compare|modifydn|delete|abandon")
	concurrent := pflag.Int("concurrent", 1, "Number of concurrent test workers")
	dryRun := pflag.Bool("dry-run", false, "Preview operations without executing")

	logLevel := pflag.String("log-level", "info", "Log level: error|warn|info|debug|trace")
	logFile := pflag.String("log-file", "", "Log file path (default: ./logs/ldap-test-{timestamp}.log)")
	verbose := pflag.BoolP("verbose", "v", false, "Enable verbose logging (sets log-level to trace)")

	cleanup := pflag.Bool("cleanup", false, "Delete test data after run")
	cleanupOnSuccess := pflag.Bool("cleanup-on-success", false, "Delete test data only if all tests pass")
	listTestData := pflag.Bool("list-test-data", false, "List existing test data and exit")
	cleanupOlderThan := pflag.String("cleanup-older-than", "", "Cleanup test data older than duration (e.g., 7d, 24h)")

	reportFormat := pflag.String("report-format", "console", "Output format: console|json|xml")
	showVersion := pflag.Bool("version", false, "Show version information")
	showHelp := pflag.BoolP("help", "h", false, "Show help message")

	pflag.Parse()

	// Show version
	if *showVersion {
		fmt.Printf("LDAP Operations Test Suite v%s\n", version)
		os.Exit(0)
	}

	// Show help
	if *showHelp {
		fmt.Println("LDAP Operations Test Suite")
		fmt.Println("\nA comprehensive testing application for validating LDAP operations.")
		fmt.Println("\nUsage:")
		fmt.Println("  ldap-test [flags]")
		fmt.Println("\nFlags:")
		pflag.PrintDefaults()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadFromFile(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override config with CLI flags (CLI flags take precedence)
	if *host != "" {
		cfg.Host = *host
	}
	if pflag.Lookup("port").Changed {
		cfg.Port = *port
	}
	if *bindDN != "" {
		cfg.BindDN = *bindDN
	}
	if *bindPassword != "" {
		cfg.BindPassword = *bindPassword
	}
	if *baseDN != "" {
		cfg.BaseDN = *baseDN
	}
	if pflag.Lookup("use-tls").Changed {
		cfg.UseTLS = *useTLS
	}
	if pflag.Lookup("start-tls").Changed {
		cfg.StartTLS = *startTLS
	}
	if pflag.Lookup("timeout").Changed {
		cfg.Timeout = *timeout
	}
	if *testPrefix != "" {
		cfg.TestPrefix = *testPrefix
	}
	if *testSuite != "" {
		cfg.TestSuite = *testSuite
	}
	if pflag.Lookup("concurrent").Changed {
		cfg.Concurrent = *concurrent
	}
	if pflag.Lookup("dry-run").Changed {
		cfg.DryRun = *dryRun
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}
	if *logFile != "" {
		cfg.LogFile = *logFile
	}
	if pflag.Lookup("verbose").Changed && *verbose {
		cfg.Verbose = true
		cfg.LogLevel = "trace"
	}
	if pflag.Lookup("cleanup").Changed {
		cfg.Cleanup = *cleanup
	}
	if pflag.Lookup("cleanup-on-success").Changed {
		cfg.CleanupOnSuccess = *cleanupOnSuccess
	}
	if pflag.Lookup("list-test-data").Changed {
		cfg.ListTestData = *listTestData
	}
	if *cleanupOlderThan != "" {
		cfg.CleanupOlderThan = *cleanupOlderThan
	}
	if *reportFormat != "" {
		cfg.ReportFormat = *reportFormat
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nRun with --help for usage information\n")
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Initialize(cfg.LogLevel, cfg.LogFile); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Main", "LDAP Operations Test Suite", "version", version)
	logger.Info("Main", "Configuration loaded", "host", cfg.Host, "port", cfg.Port, "baseDN", cfg.BaseDN)

	// Handle special modes
	if cfg.ListTestData {
		handleListTestData(cfg)
		os.Exit(0)
	}

	if cfg.CleanupOlderThan != "" {
		handleCleanupOlder(cfg)
		os.Exit(0)
	}

	// Run the test suite
	runner := tests.NewRunner(cfg)
	if err := runner.Run(); err != nil {
		logger.Error("Main", "Test suite failed", "error", err)
		fmt.Fprintf(os.Stderr, "\nTest suite failed: %v\n", err)
		os.Exit(1)
	}

	// Exit with appropriate code
	exitCode := runner.GetExitCode()
	logger.Info("Main", "Test suite completed", "exitCode", exitCode)
	os.Exit(exitCode)
}

func handleListTestData(cfg *config.Config) {
	logger.Info("Main", "Listing existing test data")
	fmt.Println("List test data functionality not yet implemented")
	// TODO: Implement listing of existing test data
	// This would require searching for entries matching the test prefix
}

func handleCleanupOlder(cfg *config.Config) {
	logger.Info("Main", "Cleaning up old test data", "olderThan", cfg.CleanupOlderThan)
	fmt.Printf("Cleanup of data older than %s not yet implemented\n", cfg.CleanupOlderThan)
	// TODO: Implement cleanup of old test data
	// This would require parsing the duration and searching for old entries
}
