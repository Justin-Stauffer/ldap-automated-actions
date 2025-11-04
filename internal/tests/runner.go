package tests

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"ldap-automated-actions/internal/config"
	"ldap-automated-actions/internal/ldap"
	"ldap-automated-actions/internal/logger"
	"ldap-automated-actions/internal/tracker"

	ldaplib "github.com/go-ldap/ldap/v3"
)

// LoopStats tracks statistics across multiple test runs
type LoopStats struct {
	TotalRuns      int
	SuccessfulRuns int
	FailedRuns     int
	TotalTests     int
	TotalPassed    int
	TotalFailed    int
	TotalDuration  time.Duration
	StartTime      time.Time
}

// Runner orchestrates the execution of all LDAP tests
type Runner struct {
	config  *config.Config
	conn    *ldap.Connection
	tracker *tracker.Tracker
	suite   *TestSuite
	loopStats *LoopStats
}

// NewRunner creates a new test runner
func NewRunner(cfg *config.Config) *Runner {
	return &Runner{
		config:  cfg,
		tracker: tracker.NewTracker(),
		suite: &TestSuite{
			Name:    "LDAP Operations Test Suite",
			Results: make([]TestResult, 0),
		},
		loopStats: &LoopStats{
			StartTime: time.Now(),
		},
	}
}

// Run executes the complete test suite
func (r *Runner) Run() error {
	// Check if loop mode is enabled
	if r.config.Loop {
		return r.RunLoop()
	}

	// Single run mode
	return r.runOnce()
}

// RunLoop executes tests continuously with statistics tracking
func (r *Runner) RunLoop() error {
	logger.Info("TestRunner", "Starting LDAP operations test suite in LOOP mode")

	if r.config.LoopCount > 0 {
		logger.Info("TestRunner", "Will run for iterations", "count", r.config.LoopCount)
	} else {
		logger.Info("TestRunner", "Running indefinitely (Ctrl+C to stop)")
	}

	if r.config.LoopDelay > 0 {
		logger.Info("TestRunner", "Delay between iterations", "seconds", r.config.LoopDelay)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	stopChan := make(chan bool)

	go func() {
		<-sigChan
		logger.Info("TestRunner", "Received interrupt signal, stopping after current iteration...")
		stopChan <- true
	}()

	iteration := 0
	for {
		iteration++

		// Check if we should stop
		select {
		case <-stopChan:
			logger.Info("TestRunner", "Stopping loop mode")
			r.reportLoopStats()
			return nil
		default:
		}

		// Check iteration limit
		if r.config.LoopCount > 0 && iteration > r.config.LoopCount {
			logger.Info("TestRunner", "Completed all iterations", "count", r.config.LoopCount)
			r.reportLoopStats()
			return nil
		}

		logger.Info("TestRunner", fmt.Sprintf("=== Starting iteration %d ===", iteration))

		// Run single test iteration
		err := r.runOnce()

		// Update statistics
		r.loopStats.TotalRuns++
		if err != nil {
			r.loopStats.FailedRuns++
			logger.Error("TestRunner", "Iteration failed", "iteration", iteration, "error", err)
		} else {
			r.loopStats.SuccessfulRuns++
		}

		// Aggregate test statistics
		total, passed, failed, duration := r.suite.GetStats()
		r.loopStats.TotalTests += total
		r.loopStats.TotalPassed += passed
		r.loopStats.TotalFailed += failed
		r.loopStats.TotalDuration += duration

		// Print iteration summary
		fmt.Printf("\n[Iteration %d] Tests: %d passed, %d failed (%.2fs)\n",
			iteration, passed, failed, duration.Seconds())

		// Print cumulative statistics
		fmt.Printf("[Cumulative] Runs: %d, Success: %d, Failed: %d, Total Tests: %d/%d (%.1f%% pass rate)\n\n",
			r.loopStats.TotalRuns,
			r.loopStats.SuccessfulRuns,
			r.loopStats.FailedRuns,
			r.loopStats.TotalPassed,
			r.loopStats.TotalTests,
			float64(r.loopStats.TotalPassed)/float64(r.loopStats.TotalTests)*100)

		// Reset suite for next iteration
		r.suite = &TestSuite{
			Name:    "LDAP Operations Test Suite",
			Results: make([]TestResult, 0),
		}
		r.tracker.Clear()

		// Delay before next iteration
		if r.config.LoopDelay > 0 {
			logger.Debug("TestRunner", "Waiting before next iteration", "seconds", r.config.LoopDelay)
			time.Sleep(time.Duration(r.config.LoopDelay) * time.Second)
		}
	}
}

// runOnce executes a single test run
func (r *Runner) runOnce() error {
	logger.Info("TestRunner", "Starting LDAP operations test suite")
	r.suite.StartTime = time.Now()

	// Phase 1: Connection and Health Check
	if err := r.connect(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer r.cleanup()

	// Phase 2: Setup (create test structure)
	testBaseDN, err := r.setup()
	if err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// Phase 3: Execute tests based on test suite selection
	r.executeTests(testBaseDN)

	// Phase 4: Cleanup (if requested)
	r.performCleanup()

	r.suite.EndTime = time.Now()

	// Phase 5: Report results (only if not in loop mode)
	if !r.config.Loop {
		r.reportResults()
	}

	return nil
}

// connect establishes connection to LDAP server
func (r *Runner) connect() error {
	logger.Info("TestRunner", "Connecting to LDAP server", "address", r.config.GetAddress())

	conn, err := ldap.NewConnection(r.config)
	if err != nil {
		logger.Error("TestRunner", "Failed to connect", "error", err)
		return err
	}
	r.conn = conn

	// Perform bind
	if err := r.conn.Bind(); err != nil {
		logger.Error("TestRunner", "Authentication failed", "error", err)
		return err
	}

	// Health check
	if err := r.conn.HealthCheck(); err != nil {
		logger.Warn("TestRunner", "Health check failed", "error", err)
	}

	return nil
}

// setup creates the test organizational structure
func (r *Runner) setup() (string, error) {
	logger.Info("Setup", "Creating test organizational structure")

	// Create timestamped test base DN
	timestamp := time.Now().Format("20060102-150405")
	testOUName := fmt.Sprintf("%s-%s", r.config.TestPrefix, timestamp)
	testBaseDN := fmt.Sprintf("ou=%s,%s", testOUName, r.config.BaseDN)

	logger.Info("Setup", "Creating test base OU", "dn", testBaseDN)

	if r.config.DryRun {
		logger.Info("Setup", "DRY RUN: Would create test base OU", "dn", testBaseDN)
		return testBaseDN, nil
	}

	// Create the test OU
	logger.Trace("Setup", "Creating test OU", "dn", testBaseDN)

	addRequest := ldaplib.NewAddRequest(testBaseDN, nil)
	addRequest.Attribute("objectClass", []string{"organizationalUnit"})
	addRequest.Attribute("ou", []string{testOUName})
	addRequest.Attribute("description", []string{fmt.Sprintf("Test OU created by LDAP test suite at %s", time.Now().Format(time.RFC3339))})

	start := time.Now()
	err := r.conn.GetConnection().Add(addRequest)
	duration := time.Since(start)

	if err != nil {
		logger.LogLDAPResult("Setup", "Add", false, -1, err.Error(), duration)
		return "", fmt.Errorf("failed to create test base OU: %w", err)
	}

	logger.LogLDAPResult("Setup", "Add", true, 0, "Success", duration)
	logger.Info("Setup", "Test OU created successfully", "dn", testBaseDN)

	// Track the test base OU
	r.tracker.Track(testBaseDN, tracker.TypeOU)

	return testBaseDN, nil
}

// executeTests runs the selected test suites
func (r *Runner) executeTests(testBaseDN string) {
	logger.Info("TestRunner", "Executing test operations", "suite", r.config.TestSuite)

	if r.config.DryRun {
		logger.Info("TestRunner", "DRY RUN: Skipping test execution")
		return
	}

	testSuite := r.config.TestSuite

	// Run tests based on suite selection
	if testSuite == "all" || testSuite == "bind" {
		results := TestBind(r.conn)
		r.suite.Results = append(r.suite.Results, results...)
	}

	if testSuite == "all" || testSuite == "add" {
		results := TestAdd(r.conn, testBaseDN, r.tracker)
		r.suite.Results = append(r.suite.Results, results...)
	}

	if testSuite == "all" || testSuite == "search" {
		results := TestSearch(r.conn, testBaseDN)
		r.suite.Results = append(r.suite.Results, results...)
	}

	if testSuite == "all" || testSuite == "compare" {
		results := TestCompare(r.conn, testBaseDN)
		r.suite.Results = append(r.suite.Results, results...)
	}

	if testSuite == "all" || testSuite == "modify" {
		results := TestModify(r.conn, testBaseDN)
		r.suite.Results = append(r.suite.Results, results...)
	}

	if testSuite == "all" || testSuite == "modifydn" {
		results := TestModifyDN(r.conn, testBaseDN, r.tracker)
		r.suite.Results = append(r.suite.Results, results...)
	}

	if testSuite == "all" || testSuite == "delete" {
		results := TestDelete(r.conn, testBaseDN, r.tracker)
		r.suite.Results = append(r.suite.Results, results...)
	}

	if testSuite == "all" || testSuite == "abandon" {
		results := TestAbandon(r.conn, r.config.BaseDN)
		r.suite.Results = append(r.suite.Results, results...)
	}

	// Note: Unbind test is run separately at the end if requested
}

// performCleanup removes test data if cleanup is enabled
func (r *Runner) performCleanup() {
	shouldCleanup := r.config.Cleanup || (r.config.CleanupOnSuccess && r.suite.AllPassed())

	if !shouldCleanup {
		logger.Info("Cleanup", "Cleanup not requested, preserving test data")
		return
	}

	if r.config.DryRun {
		logger.Info("Cleanup", "DRY RUN: Would cleanup test data")
		return
	}

	logger.Info("Cleanup", "Starting cleanup of test data")

	if err := PerformCleanup(r.conn, r.tracker); err != nil {
		logger.Warn("Cleanup", "Cleanup completed with errors", "error", err)
	} else {
		logger.Info("Cleanup", "Cleanup completed successfully")
	}
}

// cleanup closes connections and performs final operations
func (r *Runner) cleanup() {
	if r.conn != nil {
		logger.Debug("TestRunner", "Closing LDAP connection")
		r.conn.Close()
	}
}

// reportResults prints the test results
func (r *Runner) reportResults() {
	total, passed, failed, duration := r.suite.GetStats()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("LDAP OPERATIONS TEST SUITE RESULTS")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Total Tests:     %d\n", total)
	fmt.Printf("Passed:          %d\n", passed)
	fmt.Printf("Failed:          %d\n", failed)
	fmt.Printf("Duration:        %s\n", duration)
	fmt.Println(strings.Repeat("=", 80))

	// Print individual test results
	if len(r.suite.Results) > 0 {
		fmt.Println("\nDetailed Results:")
		fmt.Println(strings.Repeat("-", 80))

		currentOp := ""
		for _, result := range r.suite.Results {
			if result.Operation != currentOp {
				fmt.Printf("\n%s Tests:\n", result.Operation)
				currentOp = result.Operation
			}

			status := "✓ PASS"
			if !result.Passed {
				status = "✗ FAIL"
			}

			fmt.Printf("  %s  %-50s  %6dms\n", status, result.Name, result.Duration.Milliseconds())

			if !result.Passed && result.Error != nil {
				fmt.Printf("         Error: %v\n", result.Error)
			}
			if result.Message != "" {
				fmt.Printf("         %s\n", result.Message)
			}
		}
		fmt.Println()
	}

	// Print tracked entries summary if data was preserved
	if !r.config.Cleanup && !r.config.CleanupOnSuccess {
		r.tracker.PrintSummary()
	}

	// Overall result
	fmt.Println(strings.Repeat("=", 80))
	if r.suite.AllPassed() {
		fmt.Println("✓ ALL TESTS PASSED")
		logger.Info("TestRunner", "All tests passed")
	} else {
		fmt.Println("✗ SOME TESTS FAILED")
		logger.Warn("TestRunner", "Some tests failed", "failed", failed, "total", total)
	}
	fmt.Println(strings.Repeat("=", 80))
}

// GetExitCode returns the appropriate exit code based on test results
func (r *Runner) GetExitCode() int {
	if r.suite.AllPassed() {
		return 0
	}
	return 1
}

// reportLoopStats prints cumulative statistics from loop mode
func (r *Runner) reportLoopStats() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("LDAP OPERATIONS TEST SUITE - LOOP MODE SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	elapsed := time.Since(r.loopStats.StartTime)

	fmt.Printf("Total Runtime:        %s\n", elapsed.Round(time.Second))
	fmt.Printf("Total Iterations:     %d\n", r.loopStats.TotalRuns)
	fmt.Printf("Successful Runs:      %d (%.1f%%)\n",
		r.loopStats.SuccessfulRuns,
		float64(r.loopStats.SuccessfulRuns)/float64(r.loopStats.TotalRuns)*100)
	fmt.Printf("Failed Runs:          %d (%.1f%%)\n",
		r.loopStats.FailedRuns,
		float64(r.loopStats.FailedRuns)/float64(r.loopStats.TotalRuns)*100)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total Tests Executed: %d\n", r.loopStats.TotalTests)
	fmt.Printf("Tests Passed:         %d (%.1f%%)\n",
		r.loopStats.TotalPassed,
		float64(r.loopStats.TotalPassed)/float64(r.loopStats.TotalTests)*100)
	fmt.Printf("Tests Failed:         %d (%.1f%%)\n",
		r.loopStats.TotalFailed,
		float64(r.loopStats.TotalFailed)/float64(r.loopStats.TotalTests)*100)
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total Test Time:      %s\n", r.loopStats.TotalDuration.Round(time.Millisecond))
	fmt.Printf("Average Per Run:      %s\n", time.Duration(r.loopStats.TotalDuration.Nanoseconds()/int64(r.loopStats.TotalRuns)).Round(time.Millisecond))

	if r.loopStats.TotalTests > 0 {
		avgTestTime := time.Duration(r.loopStats.TotalDuration.Nanoseconds() / int64(r.loopStats.TotalTests))
		fmt.Printf("Average Per Test:     %s\n", avgTestTime.Round(time.Millisecond))
	}

	fmt.Println(strings.Repeat("=", 80))

	if r.loopStats.FailedRuns == 0 {
		fmt.Println("✓ ALL RUNS COMPLETED SUCCESSFULLY")
	} else {
		fmt.Printf("✗ %d RUNS FAILED\n", r.loopStats.FailedRuns)
	}

	fmt.Println(strings.Repeat("=", 80))
	logger.Info("TestRunner", "Loop mode completed", "totalRuns", r.loopStats.TotalRuns, "successful", r.loopStats.SuccessfulRuns, "failed", r.loopStats.FailedRuns)
}
