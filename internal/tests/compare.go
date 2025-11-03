package tests

import (
	"fmt"
	"time"

	"ldap-automated-actions/internal/ldap"
	"ldap-automated-actions/internal/logger"

	ldaplib "github.com/go-ldap/ldap/v3"
)

// TestCompare runs all compare operation tests
func TestCompare(conn *ldap.Connection, testBaseDN string) []TestResult {
	logger.Info("CompareTest", "Starting Compare operation tests")
	results := make([]TestResult, 0)

	// Test 1: Compare with matching value
	results = append(results, testCompareMatch(conn, testBaseDN))

	// Test 2: Compare with non-matching value
	results = append(results, testCompareNoMatch(conn, testBaseDN))

	// Test 3: Compare on non-existent entry
	results = append(results, testCompareNonExistent(conn, testBaseDN))

	// Test 4: Compare on non-existent attribute
	results = append(results, testCompareNonExistentAttribute(conn, testBaseDN))

	logger.Info("CompareTest", "Completed Compare operation tests", "total", len(results))
	return results
}

func testCompareMatch(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Compare - Matching Value Test"
	logger.Info("CompareTest", "Running: "+testName)

	dn := fmt.Sprintf("cn=testuser,%s", testBaseDN)
	attribute := "cn"
	value := "testuser"

	logger.Trace("Compare", "Operation: Compare", "dn", dn)
	logger.Trace("Compare", fmt.Sprintf("Comparing: %s = %s", attribute, value))

	start := time.Now()
	matched, err := conn.GetConnection().Compare(dn, attribute, value)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Compare",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Compare operation failed: %v", err)
		logger.LogLDAPResult("Compare", "Compare", false, -1, err.Error(), duration)
		logger.Error("CompareTest", result.Message)
	} else if matched {
		result.Passed = true
		result.Message = fmt.Sprintf("Attribute %s matches value '%s' (as expected)", attribute, value)
		logger.LogLDAPResult("Compare", "Compare", true, int(ldaplib.LDAPResultCompareTrue), "Compare True", duration)
		logger.Info("CompareTest", "PASS: "+testName, "matched", true, "duration", duration)
	} else {
		result.Passed = false
		result.Message = fmt.Sprintf("Attribute %s does not match value '%s' (unexpected)", attribute, value)
		logger.Warn("CompareTest", result.Message)
	}

	return result
}

func testCompareNoMatch(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Compare - Non-Matching Value Test"
	logger.Info("CompareTest", "Running: "+testName)

	dn := fmt.Sprintf("cn=testuser,%s", testBaseDN)
	attribute := "cn"
	value := "wrongvalue"

	logger.Trace("Compare", "Operation: Compare", "dn", dn)
	logger.Trace("Compare", fmt.Sprintf("Comparing: %s = %s", attribute, value))

	start := time.Now()
	matched, err := conn.GetConnection().Compare(dn, attribute, value)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Compare",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Compare operation failed: %v", err)
		logger.LogLDAPResult("Compare", "Compare", false, -1, err.Error(), duration)
		logger.Error("CompareTest", result.Message)
	} else if !matched {
		result.Passed = true
		result.Message = fmt.Sprintf("Attribute %s does not match value '%s' (as expected)", attribute, value)
		logger.LogLDAPResult("Compare", "Compare", true, int(ldaplib.LDAPResultCompareFalse), "Compare False", duration)
		logger.Info("CompareTest", "PASS: "+testName, "matched", false, "duration", duration)
	} else {
		result.Passed = false
		result.Message = fmt.Sprintf("Attribute %s unexpectedly matches value '%s'", attribute, value)
		logger.Warn("CompareTest", result.Message)
	}

	return result
}

func testCompareNonExistent(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Compare - Non-Existent Entry Test (Negative)"
	logger.Info("CompareTest", "Running: "+testName)

	dn := fmt.Sprintf("cn=nonexistent,%s", testBaseDN)
	attribute := "cn"
	value := "test"

	logger.Trace("Compare", "Operation: Compare (non-existent entry)", "dn", dn)

	start := time.Now()
	_, err := conn.GetConnection().Compare(dn, attribute, value)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Compare",
		Duration:  duration,
	}

	// This test SHOULD fail with "no such object" error
	if err != nil {
		if ldaplib.IsErrorWithCode(err, ldaplib.LDAPResultNoSuchObject) {
			result.Passed = true
			result.Message = "Correctly returned error for non-existent entry"
			logger.LogLDAPResult("Compare", "Compare", true, int(ldaplib.LDAPResultNoSuchObject), "No such object", duration)
			logger.Info("CompareTest", "PASS: "+testName+" (error as expected)", "duration", duration)
		} else {
			result.Passed = false
			result.Error = err
			result.Message = fmt.Sprintf("Failed with unexpected error: %v", err)
			logger.Error("CompareTest", result.Message)
		}
	} else {
		result.Passed = false
		result.Message = "ERROR: Compare succeeded on non-existent entry"
		logger.Error("CompareTest", result.Message)
	}

	return result
}

func testCompareNonExistentAttribute(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Compare - Non-Existent Attribute Test (Negative)"
	logger.Info("CompareTest", "Running: "+testName)

	dn := fmt.Sprintf("cn=testuser,%s", testBaseDN)
	attribute := "nonExistentAttribute"
	value := "test"

	logger.Trace("Compare", "Operation: Compare (non-existent attribute)", "dn", dn)
	logger.Trace("Compare", fmt.Sprintf("Comparing: %s = %s", attribute, value))

	start := time.Now()
	matched, err := conn.GetConnection().Compare(dn, attribute, value)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Compare",
		Duration:  duration,
	}

	if err != nil {
		// Some servers return an error for non-existent attributes
		result.Passed = true
		result.Message = "Correctly returned error for non-existent attribute"
		logger.LogLDAPResult("Compare", "Compare", true, -1, err.Error(), duration)
		logger.Info("CompareTest", "PASS: "+testName+" (error as expected)", "duration", duration)
	} else if !matched {
		// Some servers return false for non-existent attributes
		result.Passed = true
		result.Message = "Correctly returned false for non-existent attribute"
		logger.LogLDAPResult("Compare", "Compare", true, int(ldaplib.LDAPResultCompareFalse), "Compare False", duration)
		logger.Info("CompareTest", "PASS: "+testName+" (false as expected)", "duration", duration)
	} else {
		result.Passed = false
		result.Message = "ERROR: Compare returned true for non-existent attribute"
		logger.Error("CompareTest", result.Message)
	}

	return result
}
