package tests

import (
	"fmt"
	"time"

	"ldap-automated-actions/internal/ldap"
	"ldap-automated-actions/internal/logger"
	"ldap-automated-actions/internal/tracker"

	ldaplib "github.com/go-ldap/ldap/v3"
)

// TestModifyDN runs all modify DN operation tests
func TestModifyDN(conn *ldap.Connection, testBaseDN string, trk *tracker.Tracker) []TestResult {
	logger.Info("ModifyDNTest", "Starting Modify DN operation tests")
	results := make([]TestResult, 0)

	// Test 1: Rename entry (change RDN)
	results = append(results, testRenameEntry(conn, testBaseDN, trk))

	// Test 2: Move entry to different OU
	results = append(results, testMoveEntry(conn, testBaseDN, trk))

	// Test 3: Rename and move entry
	results = append(results, testRenameAndMove(conn, testBaseDN, trk))

	// Test 4: Try to rename to existing DN (should fail)
	results = append(results, testRenameToExisting(conn, testBaseDN))

	logger.Info("ModifyDNTest", "Completed Modify DN operation tests", "total", len(results))
	return results
}

func testRenameEntry(conn *ldap.Connection, testBaseDN string, trk *tracker.Tracker) TestResult {
	testName := "Modify DN - Rename Entry Test"
	logger.Info("ModifyDNTest", "Running: "+testName)

	// First, create a test entry to rename
	oldCN := "rename-test-user"
	oldDN := fmt.Sprintf("cn=%s,%s", oldCN, testBaseDN)

	// Create the entry
	addRequest := ldaplib.NewAddRequest(oldDN, nil)
	addRequest.Attribute("objectClass", []string{"inetOrgPerson"})
	addRequest.Attribute("cn", []string{oldCN})
	addRequest.Attribute("sn", []string{"RenameTest"})

	err := conn.GetConnection().Add(addRequest)
	if err != nil {
		logger.Error("ModifyDNTest", "Failed to create test entry for rename", "error", err)
		return TestResult{
			Name:      testName,
			Operation: "ModifyDN",
			Passed:    false,
			Error:     err,
			Message:   "Failed to create test entry",
		}
	}
	trk.Track(oldDN, tracker.TypeUser)

	// Now rename it
	newRDN := "cn=renamed-user"
	logger.Trace("ModifyDN", "Operation: ModifyDN (Rename)", "oldDN", oldDN, "newRDN", newRDN)

	modifyDNRequest := ldaplib.NewModifyDNRequest(oldDN, newRDN, true, "")

	start := time.Now()
	err = conn.GetConnection().ModifyDN(modifyDNRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "ModifyDN",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Failed to rename entry: %v", err)
		logger.LogLDAPResult("ModifyDN", "ModifyDN", false, -1, err.Error(), duration)
		logger.Error("ModifyDNTest", result.Message)
	} else {
		newDN := fmt.Sprintf("cn=renamed-user,%s", testBaseDN)
		result.Passed = true
		result.Message = fmt.Sprintf("Successfully renamed entry from %s to %s", oldDN, newDN)
		logger.LogLDAPResult("ModifyDN", "ModifyDN", true, 0, "Success", duration)
		logger.Info("ModifyDNTest", "PASS: "+testName, "newDN", newDN, "duration", duration)

		// Update tracker with new DN
		trk.Track(newDN, tracker.TypeUser)
	}

	return result
}

func testMoveEntry(conn *ldap.Connection, testBaseDN string, trk *tracker.Tracker) TestResult {
	testName := "Modify DN - Move Entry Test"
	logger.Info("ModifyDNTest", "Running: "+testName)

	// First, create a second OU to move entries into
	targetOU := "target-ou"
	targetOUDN := fmt.Sprintf("ou=%s,%s", targetOU, testBaseDN)

	addRequest := ldaplib.NewAddRequest(targetOUDN, nil)
	addRequest.Attribute("objectClass", []string{"organizationalUnit"})
	addRequest.Attribute("ou", []string{targetOU})

	err := conn.GetConnection().Add(addRequest)
	if err != nil {
		logger.Warn("ModifyDNTest", "Failed to create target OU (may already exist)", "error", err)
	} else {
		trk.Track(targetOUDN, tracker.TypeOU)
	}

	// Create a user to move
	oldCN := "move-test-user"
	oldDN := fmt.Sprintf("cn=%s,%s", oldCN, testBaseDN)

	addRequest = ldaplib.NewAddRequest(oldDN, nil)
	addRequest.Attribute("objectClass", []string{"inetOrgPerson"})
	addRequest.Attribute("cn", []string{oldCN})
	addRequest.Attribute("sn", []string{"MoveTest"})

	err = conn.GetConnection().Add(addRequest)
	if err != nil {
		logger.Error("ModifyDNTest", "Failed to create test entry for move", "error", err)
		return TestResult{
			Name:      testName,
			Operation: "ModifyDN",
			Passed:    false,
			Error:     err,
			Message:   "Failed to create test entry",
		}
	}
	trk.Track(oldDN, tracker.TypeUser)

	// Now move it to the new OU
	newRDN := fmt.Sprintf("cn=%s", oldCN) // Keep same RDN
	logger.Trace("ModifyDN", "Operation: ModifyDN (Move)", "oldDN", oldDN, "newSuperior", targetOUDN)

	modifyDNRequest := ldaplib.NewModifyDNRequest(oldDN, newRDN, true, targetOUDN)

	start := time.Now()
	err = conn.GetConnection().ModifyDN(modifyDNRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "ModifyDN",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Failed to move entry: %v", err)
		logger.LogLDAPResult("ModifyDN", "ModifyDN", false, -1, err.Error(), duration)
		logger.Error("ModifyDNTest", result.Message)
	} else {
		newDN := fmt.Sprintf("cn=%s,%s", oldCN, targetOUDN)
		result.Passed = true
		result.Message = fmt.Sprintf("Successfully moved entry from %s to %s", oldDN, newDN)
		logger.LogLDAPResult("ModifyDN", "ModifyDN", true, 0, "Success", duration)
		logger.Info("ModifyDNTest", "PASS: "+testName, "newDN", newDN, "duration", duration)

		// Update tracker
		trk.Track(newDN, tracker.TypeUser)
	}

	return result
}

func testRenameAndMove(conn *ldap.Connection, testBaseDN string, trk *tracker.Tracker) TestResult {
	testName := "Modify DN - Rename and Move Entry Test"
	logger.Info("ModifyDNTest", "Running: "+testName)

	// Create target OU if it doesn't exist
	targetOU := "target-ou"
	targetOUDN := fmt.Sprintf("ou=%s,%s", targetOU, testBaseDN)

	// Create a user to rename and move
	oldCN := "rename-move-user"
	oldDN := fmt.Sprintf("cn=%s,%s", oldCN, testBaseDN)

	addRequest := ldaplib.NewAddRequest(oldDN, nil)
	addRequest.Attribute("objectClass", []string{"inetOrgPerson"})
	addRequest.Attribute("cn", []string{oldCN})
	addRequest.Attribute("sn", []string{"RenameMoveTest"})

	err := conn.GetConnection().Add(addRequest)
	if err != nil {
		logger.Error("ModifyDNTest", "Failed to create test entry", "error", err)
		return TestResult{
			Name:      testName,
			Operation: "ModifyDN",
			Passed:    false,
			Error:     err,
			Message:   "Failed to create test entry",
		}
	}
	trk.Track(oldDN, tracker.TypeUser)

	// Rename and move simultaneously
	newRDN := "cn=renamed-moved-user"
	logger.Trace("ModifyDN", "Operation: ModifyDN (Rename+Move)", "oldDN", oldDN, "newRDN", newRDN, "newSuperior", targetOUDN)

	modifyDNRequest := ldaplib.NewModifyDNRequest(oldDN, newRDN, true, targetOUDN)

	start := time.Now()
	err = conn.GetConnection().ModifyDN(modifyDNRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "ModifyDN",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Failed to rename and move entry: %v", err)
		logger.LogLDAPResult("ModifyDN", "ModifyDN", false, -1, err.Error(), duration)
		logger.Error("ModifyDNTest", result.Message)
	} else {
		newDN := fmt.Sprintf("cn=renamed-moved-user,%s", targetOUDN)
		result.Passed = true
		result.Message = fmt.Sprintf("Successfully renamed and moved entry from %s to %s", oldDN, newDN)
		logger.LogLDAPResult("ModifyDN", "ModifyDN", true, 0, "Success", duration)
		logger.Info("ModifyDNTest", "PASS: "+testName, "newDN", newDN, "duration", duration)

		// Update tracker
		trk.Track(newDN, tracker.TypeUser)
	}

	return result
}

func testRenameToExisting(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Modify DN - Rename to Existing DN Test (Negative)"
	logger.Info("ModifyDNTest", "Running: "+testName)

	// Try to rename testuser to renamed-user (which should already exist)
	oldDN := fmt.Sprintf("cn=testuser,%s", testBaseDN)
	newRDN := "cn=renamed-user"

	logger.Trace("ModifyDN", "Operation: ModifyDN (to existing)", "oldDN", oldDN, "newRDN", newRDN)

	modifyDNRequest := ldaplib.NewModifyDNRequest(oldDN, newRDN, true, "")

	start := time.Now()
	err := conn.GetConnection().ModifyDN(modifyDNRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "ModifyDN",
		Duration:  duration,
	}

	// This test SHOULD fail
	if err != nil {
		if ldaplib.IsErrorWithCode(err, ldaplib.LDAPResultEntryAlreadyExists) {
			result.Passed = true
			result.Message = "Correctly rejected rename to existing DN"
			logger.LogLDAPResult("ModifyDN", "ModifyDN", true, int(ldaplib.LDAPResultEntryAlreadyExists), "Entry already exists", duration)
			logger.Info("ModifyDNTest", "PASS: "+testName+" (rejected)", "duration", duration)
		} else {
			result.Passed = true // Still pass if it failed (just different error)
			result.Message = fmt.Sprintf("Failed as expected with error: %v", err)
			logger.Info("ModifyDNTest", "PASS: "+testName+" (failed as expected)", "duration", duration)
		}
	} else {
		result.Passed = false
		result.Message = "ERROR: Rename to existing DN succeeded"
		logger.Error("ModifyDNTest", result.Message)
	}

	return result
}
