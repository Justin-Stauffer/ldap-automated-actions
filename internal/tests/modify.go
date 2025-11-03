package tests

import (
	"fmt"
	"time"

	"ldap-automated-actions/internal/ldap"
	"ldap-automated-actions/internal/logger"

	ldaplib "github.com/go-ldap/ldap/v3"
)

// TestModify runs all modify operation tests
func TestModify(conn *ldap.Connection, testBaseDN string) []TestResult {
	logger.Info("ModifyTest", "Starting Modify operation tests")
	results := make([]TestResult, 0)

	// Test 1: Add attribute value
	results = append(results, testModifyAddAttribute(conn, testBaseDN))

	// Test 2: Replace attribute value
	results = append(results, testModifyReplaceAttribute(conn, testBaseDN))

	// Test 3: Delete attribute value
	results = append(results, testModifyDeleteAttribute(conn, testBaseDN))

	// Test 4: Multiple modifications in one request
	results = append(results, testModifyMultiple(conn, testBaseDN))

	// Test 5: Modify non-existent entry (should fail)
	results = append(results, testModifyNonExistent(conn, testBaseDN))

	logger.Info("ModifyTest", "Completed Modify operation tests", "total", len(results))
	return results
}

func testModifyAddAttribute(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Modify - Add Attribute Test"
	logger.Info("ModifyTest", "Running: "+testName)

	dn := fmt.Sprintf("cn=testuser,%s", testBaseDN)

	modifyRequest := ldaplib.NewModifyRequest(dn, nil)
	modifyRequest.Add("telephoneNumber", []string{"+1-555-0100"})

	logger.Trace("Modify", "Operation: Modify (Add)", "dn", dn)
	logger.Trace("Modify", fmt.Sprintf("Adding attribute: telephoneNumber = +1-555-0100"))

	start := time.Now()
	err := conn.GetConnection().Modify(modifyRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Modify",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Failed to add attribute: %v", err)
		logger.LogLDAPResult("Modify", "Modify (Add)", false, -1, err.Error(), duration)
		logger.Error("ModifyTest", result.Message)
	} else {
		result.Passed = true
		result.Message = "Successfully added telephoneNumber attribute"
		logger.LogLDAPResult("Modify", "Modify (Add)", true, 0, "Success", duration)
		logger.Info("ModifyTest", "PASS: "+testName, "dn", dn, "duration", duration)
	}

	return result
}

func testModifyReplaceAttribute(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Modify - Replace Attribute Test"
	logger.Info("ModifyTest", "Running: "+testName)

	dn := fmt.Sprintf("cn=testuser,%s", testBaseDN)

	modifyRequest := ldaplib.NewModifyRequest(dn, nil)
	modifyRequest.Replace("mail", []string{"newemail@example.com"})

	logger.Trace("Modify", "Operation: Modify (Replace)", "dn", dn)
	logger.Trace("Modify", fmt.Sprintf("Replacing attribute: mail = newemail@example.com"))

	start := time.Now()
	err := conn.GetConnection().Modify(modifyRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Modify",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Failed to replace attribute: %v", err)
		logger.LogLDAPResult("Modify", "Modify (Replace)", false, -1, err.Error(), duration)
		logger.Error("ModifyTest", result.Message)
	} else {
		result.Passed = true
		result.Message = "Successfully replaced mail attribute"
		logger.LogLDAPResult("Modify", "Modify (Replace)", true, 0, "Success", duration)
		logger.Info("ModifyTest", "PASS: "+testName, "dn", dn, "duration", duration)
	}

	return result
}

func testModifyDeleteAttribute(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Modify - Delete Attribute Test"
	logger.Info("ModifyTest", "Running: "+testName)

	dn := fmt.Sprintf("cn=testuser,%s", testBaseDN)

	modifyRequest := ldaplib.NewModifyRequest(dn, nil)
	modifyRequest.Delete("telephoneNumber", []string{}) // Delete all values

	logger.Trace("Modify", "Operation: Modify (Delete)", "dn", dn)
	logger.Trace("Modify", fmt.Sprintf("Deleting attribute: telephoneNumber"))

	start := time.Now()
	err := conn.GetConnection().Modify(modifyRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Modify",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Failed to delete attribute: %v", err)
		logger.LogLDAPResult("Modify", "Modify (Delete)", false, -1, err.Error(), duration)
		logger.Error("ModifyTest", result.Message)
	} else {
		result.Passed = true
		result.Message = "Successfully deleted telephoneNumber attribute"
		logger.LogLDAPResult("Modify", "Modify (Delete)", true, 0, "Success", duration)
		logger.Info("ModifyTest", "PASS: "+testName, "dn", dn, "duration", duration)
	}

	return result
}

func testModifyMultiple(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Modify - Multiple Modifications Test"
	logger.Info("ModifyTest", "Running: "+testName)

	dn := fmt.Sprintf("cn=testuser,%s", testBaseDN)

	modifyRequest := ldaplib.NewModifyRequest(dn, nil)
	modifyRequest.Add("mobile", []string{"+1-555-0200"})
	modifyRequest.Replace("description", []string{"Modified test user with multiple changes"})

	logger.Trace("Modify", "Operation: Modify (Multiple)", "dn", dn)
	logger.Trace("Modify", "Modifications: Add mobile, Replace description")

	start := time.Now()
	err := conn.GetConnection().Modify(modifyRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Modify",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Failed to apply multiple modifications: %v", err)
		logger.LogLDAPResult("Modify", "Modify (Multiple)", false, -1, err.Error(), duration)
		logger.Error("ModifyTest", result.Message)
	} else {
		result.Passed = true
		result.Message = "Successfully applied multiple modifications"
		logger.LogLDAPResult("Modify", "Modify (Multiple)", true, 0, "Success", duration)
		logger.Info("ModifyTest", "PASS: "+testName, "dn", dn, "duration", duration)
	}

	return result
}

func testModifyNonExistent(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Modify - Non-Existent Entry Test (Negative)"
	logger.Info("ModifyTest", "Running: "+testName)

	dn := fmt.Sprintf("cn=nonexistent,%s", testBaseDN)

	modifyRequest := ldaplib.NewModifyRequest(dn, nil)
	modifyRequest.Replace("description", []string{"This should fail"})

	logger.Trace("Modify", "Operation: Modify (non-existent)", "dn", dn)

	start := time.Now()
	err := conn.GetConnection().Modify(modifyRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Modify",
		Duration:  duration,
	}

	// This test SHOULD fail - we expect an error
	if err != nil {
		if ldaplib.IsErrorWithCode(err, ldaplib.LDAPResultNoSuchObject) {
			result.Passed = true
			result.Message = "Correctly rejected modification of non-existent entry"
			logger.LogLDAPResult("Modify", "Modify", true, int(ldaplib.LDAPResultNoSuchObject), "No such object", duration)
			logger.Info("ModifyTest", "PASS: "+testName+" (rejected)", "duration", duration)
		} else {
			result.Passed = false
			result.Error = err
			result.Message = fmt.Sprintf("Failed with unexpected error: %v", err)
			logger.Error("ModifyTest", result.Message)
		}
	} else {
		result.Passed = false
		result.Message = "ERROR: Modification of non-existent entry succeeded"
		logger.Error("ModifyTest", result.Message)
	}

	return result
}
