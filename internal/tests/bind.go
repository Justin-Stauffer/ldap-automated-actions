package tests

import (
	"fmt"
	"time"

	"ldap-automated-actions/internal/ldap"
	"ldap-automated-actions/internal/logger"

	ldaplib "github.com/go-ldap/ldap/v3"
)

// TestBind runs all bind operation tests
func TestBind(conn *ldap.Connection) []TestResult {
	logger.Info("BindTest", "Starting Bind operation tests")
	results := make([]TestResult, 0)

	// Test 1: Valid bind (already done during connection, but test again)
	results = append(results, testValidBind(conn))

	// Test 2: Invalid credentials bind
	results = append(results, testInvalidBind(conn))

	// Test 3: Anonymous bind (if supported)
	results = append(results, testAnonymousBind(conn))

	logger.Info("BindTest", "Completed Bind operation tests", "total", len(results))
	return results
}

func testValidBind(conn *ldap.Connection) TestResult {
	testName := "Valid Bind Test"
	logger.Info("BindTest", "Running: "+testName)

	start := time.Now()
	err := conn.Bind()
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Bind",
		Duration:  duration,
		Passed:    err == nil,
	}

	if err != nil {
		result.Error = err
		result.Message = fmt.Sprintf("Failed to bind with valid credentials: %v", err)
		logger.Error("BindTest", result.Message)
	} else {
		result.Message = "Successfully authenticated with valid credentials"
		logger.Info("BindTest", "PASS: "+testName, "duration", duration)
	}

	return result
}

func testInvalidBind(conn *ldap.Connection) TestResult {
	testName := "Invalid Bind Test"
	logger.Info("BindTest", "Running: "+testName)

	cfg := conn.GetConfig()

	// Create a new connection for this test
	start := time.Now()

	// Try to dial and bind with invalid password
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	testConn, err := ldaplib.Dial("tcp", address)
	if err != nil {
		duration := time.Since(start)
		logger.Error("BindTest", "Failed to connect for invalid bind test", "error", err)
		return TestResult{
			Name:      testName,
			Operation: "Bind",
			Duration:  duration,
			Passed:    false,
			Error:     err,
			Message:   "Failed to connect to server for test",
		}
	}
	defer testConn.Close()

	// Attempt bind with invalid password
	invalidPassword := "INVALID_PASSWORD_12345"
	logger.Debug("BindTest", "Attempting bind with invalid credentials", "dn", cfg.BindDN)

	err = testConn.Bind(cfg.BindDN, invalidPassword)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Bind",
		Duration:  duration,
	}

	// This test SHOULD fail - we expect an error
	if err != nil {
		// Check if it's an invalid credentials error
		if ldaplib.IsErrorWithCode(err, ldaplib.LDAPResultInvalidCredentials) {
			result.Passed = true
			result.Message = "Correctly rejected invalid credentials"
			logger.Info("BindTest", "PASS: "+testName+" (invalid credentials rejected)", "duration", duration)
		} else {
			result.Passed = true // Still a pass as bind failed (different error)
			result.Message = fmt.Sprintf("Bind failed as expected (error: %v)", err)
			logger.Info("BindTest", "PASS: "+testName+" (bind failed as expected)", "duration", duration)
		}
	} else {
		result.Passed = false
		result.Message = "ERROR: Invalid credentials were accepted (security issue!)"
		logger.Error("BindTest", result.Message)
	}

	return result
}

func testAnonymousBind(conn *ldap.Connection) TestResult {
	testName := "Anonymous Bind Test"
	logger.Info("BindTest", "Running: "+testName)

	cfg := conn.GetConfig()

	// Create a new connection for this test
	start := time.Now()

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	testConn, err := ldaplib.Dial("tcp", address)
	if err != nil {
		duration := time.Since(start)
		logger.Error("BindTest", "Failed to connect for anonymous bind test", "error", err)
		return TestResult{
			Name:      testName,
			Operation: "Bind",
			Duration:  duration,
			Passed:    false,
			Error:     err,
			Message:   "Failed to connect to server for test",
		}
	}
	defer testConn.Close()

	// Attempt anonymous bind (empty DN and password)
	logger.Debug("BindTest", "Attempting anonymous bind")
	err = testConn.Bind("", "")
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Bind",
		Duration:  duration,
	}

	if err != nil {
		// Anonymous bind not allowed - this is acceptable
		result.Passed = true
		result.Message = "Anonymous bind not permitted (as expected)"
		logger.Info("BindTest", "PASS: "+testName+" (anonymous bind rejected)", "duration", duration)
	} else {
		// Anonymous bind succeeded - test passes (some servers allow this)
		result.Passed = true
		result.Message = "Anonymous bind permitted on this server"
		logger.Info("BindTest", "PASS: "+testName+" (anonymous bind allowed)", "duration", duration)
	}

	return result
}
