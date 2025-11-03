package tests

import (
	"fmt"
	"time"

	"ldap-automated-actions/internal/ldap"
	"ldap-automated-actions/internal/logger"
	"ldap-automated-actions/internal/tracker"

	ldaplib "github.com/go-ldap/ldap/v3"
)

// TestAdd runs all add operation tests
func TestAdd(conn *ldap.Connection, testBaseDN string, trk *tracker.Tracker) []TestResult {
	logger.Info("AddTest", "Starting Add operation tests")
	results := make([]TestResult, 0)

	// Test 1: Add an OU
	results = append(results, testAddOU(conn, testBaseDN, trk))

	// Test 2: Add a user
	results = append(results, testAddUser(conn, testBaseDN, trk))

	// Test 3: Add a group
	results = append(results, testAddGroup(conn, testBaseDN, trk))

	// Test 4: Try to add duplicate entry (should fail)
	results = append(results, testAddDuplicate(conn, testBaseDN))

	// Test 5: Try to add entry with missing required attributes
	results = append(results, testAddMissingAttributes(conn, testBaseDN))

	logger.Info("AddTest", "Completed Add operation tests", "total", len(results))
	return results
}

func testAddOU(conn *ldap.Connection, testBaseDN string, trk *tracker.Tracker) TestResult {
	testName := "Add OU Test"
	logger.Info("AddTest", "Running: "+testName)

	ouName := "test-ou"
	dn := fmt.Sprintf("ou=%s,%s", ouName, testBaseDN)

	attributes := map[string][]string{
		"objectClass": {"organizationalUnit"},
		"ou":          {ouName},
		"description": {"Test organizational unit created by automated tests"},
	}

	start := time.Now()
	logger.Trace("Add", "Operation: Add", "dn", dn)
	logger.Trace("Add", "DN: "+dn)
	logger.Trace("Add", fmt.Sprintf("Attributes: %v", attributes))

	addRequest := ldaplib.NewAddRequest(dn, nil)
	for attr, values := range attributes {
		addRequest.Attribute(attr, values)
	}

	err := conn.GetConnection().Add(addRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Add",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Failed to add OU: %v", err)
		logger.LogLDAPResult("Add", "Add", false, -1, err.Error(), duration)
		logger.Error("AddTest", result.Message)
	} else {
		result.Passed = true
		result.Message = fmt.Sprintf("Successfully added OU: %s", dn)
		logger.LogLDAPResult("Add", "Add", true, 0, "Success", duration)
		logger.Info("AddTest", "PASS: "+testName, "dn", dn, "duration", duration)

		// Track the created entry
		trk.Track(dn, tracker.TypeOU)
	}

	return result
}

func testAddUser(conn *ldap.Connection, testBaseDN string, trk *tracker.Tracker) TestResult {
	testName := "Add User Test"
	logger.Info("AddTest", "Running: "+testName)

	cn := "testuser"
	dn := fmt.Sprintf("cn=%s,%s", cn, testBaseDN)

	attributes := map[string][]string{
		"objectClass": {"inetOrgPerson"},
		"cn":          {cn},
		"sn":          {"User"},
		"givenName":   {"Test"},
		"mail":        {"testuser@example.com"},
		"userPassword": {"TestPassword123!"},
		"description": {"Test user created by automated tests"},
	}

	start := time.Now()
	logger.Trace("Add", "Operation: Add", "dn", dn)
	logger.Trace("Add", "DN: "+dn)
	logger.Trace("Add", fmt.Sprintf("Attributes: %v", attributes))

	addRequest := ldaplib.NewAddRequest(dn, nil)
	for attr, values := range attributes {
		addRequest.Attribute(attr, values)
	}

	err := conn.GetConnection().Add(addRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Add",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Failed to add user: %v", err)
		logger.LogLDAPResult("Add", "Add", false, -1, err.Error(), duration)
		logger.Error("AddTest", result.Message)
	} else {
		result.Passed = true
		result.Message = fmt.Sprintf("Successfully added user: %s", dn)
		logger.LogLDAPResult("Add", "Add", true, 0, "Success", duration)
		logger.Info("AddTest", "PASS: "+testName, "dn", dn, "duration", duration)

		// Track the created entry
		trk.Track(dn, tracker.TypeUser)
	}

	return result
}

func testAddGroup(conn *ldap.Connection, testBaseDN string, trk *tracker.Tracker) TestResult {
	testName := "Add Group Test"
	logger.Info("AddTest", "Running: "+testName)

	cn := "testgroup"
	dn := fmt.Sprintf("cn=%s,%s", cn, testBaseDN)

	attributes := map[string][]string{
		"objectClass": {"groupOfNames"},
		"cn":          {cn},
		"description": {"Test group created by automated tests"},
		"member":      {fmt.Sprintf("cn=testuser,%s", testBaseDN)}, // Reference the user we created
	}

	start := time.Now()
	logger.Trace("Add", "Operation: Add", "dn", dn)
	logger.Trace("Add", "DN: "+dn)
	logger.Trace("Add", fmt.Sprintf("Attributes: %v", attributes))

	addRequest := ldaplib.NewAddRequest(dn, nil)
	for attr, values := range attributes {
		addRequest.Attribute(attr, values)
	}

	err := conn.GetConnection().Add(addRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Add",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Failed to add group: %v", err)
		logger.LogLDAPResult("Add", "Add", false, -1, err.Error(), duration)
		logger.Error("AddTest", result.Message)
	} else {
		result.Passed = true
		result.Message = fmt.Sprintf("Successfully added group: %s", dn)
		logger.LogLDAPResult("Add", "Add", true, 0, "Success", duration)
		logger.Info("AddTest", "PASS: "+testName, "dn", dn, "duration", duration)

		// Track the created entry
		trk.Track(dn, tracker.TypeGroup)
	}

	return result
}

func testAddDuplicate(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Add Duplicate Entry Test (Negative)"
	logger.Info("AddTest", "Running: "+testName)

	// Try to add the same user again
	cn := "testuser"
	dn := fmt.Sprintf("cn=%s,%s", cn, testBaseDN)

	attributes := map[string][]string{
		"objectClass": {"inetOrgPerson"},
		"cn":          {cn},
		"sn":          {"User"},
	}

	start := time.Now()
	logger.Trace("Add", "Operation: Add (duplicate)", "dn", dn)

	addRequest := ldaplib.NewAddRequest(dn, nil)
	for attr, values := range attributes {
		addRequest.Attribute(attr, values)
	}

	err := conn.GetConnection().Add(addRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Add",
		Duration:  duration,
	}

	// This test SHOULD fail - we expect an error
	if err != nil {
		if ldaplib.IsErrorWithCode(err, ldaplib.LDAPResultEntryAlreadyExists) {
			result.Passed = true
			result.Message = "Correctly rejected duplicate entry"
			logger.LogLDAPResult("Add", "Add", true, int(ldaplib.LDAPResultEntryAlreadyExists), "Entry already exists", duration)
			logger.Info("AddTest", "PASS: "+testName+" (duplicate rejected)", "duration", duration)
		} else {
			result.Passed = false
			result.Error = err
			result.Message = fmt.Sprintf("Failed with unexpected error: %v", err)
			logger.Error("AddTest", result.Message)
		}
	} else {
		result.Passed = false
		result.Message = "ERROR: Duplicate entry was accepted"
		logger.Error("AddTest", result.Message)
	}

	return result
}

func testAddMissingAttributes(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Add Entry with Missing Required Attributes Test (Negative)"
	logger.Info("AddTest", "Running: "+testName)

	cn := "incomplete-user"
	dn := fmt.Sprintf("cn=%s,%s", cn, testBaseDN)

	// Missing required 'sn' attribute for inetOrgPerson
	attributes := map[string][]string{
		"objectClass": {"inetOrgPerson"},
		"cn":          {cn},
		// Missing sn (surname) - required attribute
	}

	start := time.Now()
	logger.Trace("Add", "Operation: Add (missing attributes)", "dn", dn)

	addRequest := ldaplib.NewAddRequest(dn, nil)
	for attr, values := range attributes {
		addRequest.Attribute(attr, values)
	}

	err := conn.GetConnection().Add(addRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Add",
		Duration:  duration,
	}

	// This test SHOULD fail - we expect an error
	if err != nil {
		result.Passed = true
		result.Message = "Correctly rejected entry with missing required attributes"
		logger.LogLDAPResult("Add", "Add", true, -1, "Missing required attributes", duration)
		logger.Info("AddTest", "PASS: "+testName+" (rejected)", "duration", duration)
	} else {
		result.Passed = false
		result.Message = "ERROR: Entry with missing required attributes was accepted"
		logger.Error("AddTest", result.Message)
	}

	return result
}
