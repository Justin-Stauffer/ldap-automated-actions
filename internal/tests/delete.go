package tests

import (
	"fmt"
	"time"

	"ldap-automated-actions/internal/ldap"
	"ldap-automated-actions/internal/logger"
	"ldap-automated-actions/internal/tracker"

	ldaplib "github.com/go-ldap/ldap/v3"
)

// TestDelete runs all delete operation tests
func TestDelete(conn *ldap.Connection, testBaseDN string, trk *tracker.Tracker) []TestResult {
	logger.Info("DeleteTest", "Starting Delete operation tests")
	results := make([]TestResult, 0)

	// Test 1: Delete a leaf entry
	results = append(results, testDeleteLeaf(conn, testBaseDN, trk))

	// Test 2: Try to delete non-leaf entry (should fail)
	results = append(results, testDeleteNonLeaf(conn, testBaseDN))

	// Test 3: Try to delete non-existent entry (should fail)
	results = append(results, testDeleteNonExistent(conn, testBaseDN))

	logger.Info("DeleteTest", "Completed Delete operation tests", "total", len(results))
	return results
}

func testDeleteLeaf(conn *ldap.Connection, testBaseDN string, trk *tracker.Tracker) TestResult {
	testName := "Delete - Leaf Entry Test"
	logger.Info("DeleteTest", "Running: "+testName)

	// Create a temporary user to delete
	cn := "delete-test-user"
	dn := fmt.Sprintf("cn=%s,%s", cn, testBaseDN)

	// Create the entry
	addRequest := ldaplib.NewAddRequest(dn, nil)
	addRequest.Attribute("objectClass", []string{"inetOrgPerson"})
	addRequest.Attribute("cn", []string{cn})
	addRequest.Attribute("sn", []string{"DeleteTest"})

	err := conn.GetConnection().Add(addRequest)
	if err != nil {
		logger.Error("DeleteTest", "Failed to create test entry for deletion", "error", err)
		return TestResult{
			Name:      testName,
			Operation: "Delete",
			Passed:    false,
			Error:     err,
			Message:   "Failed to create test entry",
		}
	}

	logger.Debug("DeleteTest", "Created temporary entry for deletion", "dn", dn)

	// Now delete it
	logger.Trace("Delete", "Operation: Delete", "dn", dn)

	delRequest := ldaplib.NewDelRequest(dn, nil)

	start := time.Now()
	err = conn.GetConnection().Del(delRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Delete",
		Duration:  duration,
	}

	if err != nil {
		result.Passed = false
		result.Error = err
		result.Message = fmt.Sprintf("Failed to delete entry: %v", err)
		logger.LogLDAPResult("Delete", "Delete", false, -1, err.Error(), duration)
		logger.Error("DeleteTest", result.Message)
		// Entry still exists, so track it for cleanup
		trk.Track(dn, tracker.TypeUser)
	} else {
		result.Passed = true
		result.Message = fmt.Sprintf("Successfully deleted entry: %s", dn)
		logger.LogLDAPResult("Delete", "Delete", true, 0, "Success", duration)
		logger.Info("DeleteTest", "PASS: "+testName, "dn", dn, "duration", duration)
		// Entry was deleted, no need to track
	}

	return result
}

func testDeleteNonLeaf(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Delete - Non-Leaf Entry Test (Negative)"
	logger.Info("DeleteTest", "Running: "+testName)

	// Try to delete the test base DN which should have child entries
	dn := testBaseDN

	logger.Trace("Delete", "Operation: Delete (non-leaf)", "dn", dn)

	delRequest := ldaplib.NewDelRequest(dn, nil)

	start := time.Now()
	err := conn.GetConnection().Del(delRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Delete",
		Duration:  duration,
	}

	// This test SHOULD fail with "not allowed on non-leaf" error
	if err != nil {
		if ldaplib.IsErrorWithCode(err, ldaplib.LDAPResultNotAllowedOnNonLeaf) {
			result.Passed = true
			result.Message = "Correctly rejected deletion of non-leaf entry"
			logger.LogLDAPResult("Delete", "Delete", true, int(ldaplib.LDAPResultNotAllowedOnNonLeaf), "Not allowed on non-leaf", duration)
			logger.Info("DeleteTest", "PASS: "+testName+" (rejected)", "duration", duration)
		} else {
			// Some other error is also acceptable (server might return different error codes)
			result.Passed = true
			result.Message = fmt.Sprintf("Correctly rejected with error: %v", err)
			logger.Info("DeleteTest", "PASS: "+testName+" (rejected with error)", "duration", duration)
		}
	} else {
		result.Passed = false
		result.Message = "ERROR: Deletion of non-leaf entry succeeded"
		logger.Error("DeleteTest", result.Message)
	}

	return result
}

func testDeleteNonExistent(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Delete - Non-Existent Entry Test (Negative)"
	logger.Info("DeleteTest", "Running: "+testName)

	dn := fmt.Sprintf("cn=nonexistent-delete-test,%s", testBaseDN)

	logger.Trace("Delete", "Operation: Delete (non-existent)", "dn", dn)

	delRequest := ldaplib.NewDelRequest(dn, nil)

	start := time.Now()
	err := conn.GetConnection().Del(delRequest)
	duration := time.Since(start)

	result := TestResult{
		Name:      testName,
		Operation: "Delete",
		Duration:  duration,
	}

	// This test SHOULD fail with "no such object" error
	if err != nil {
		if ldaplib.IsErrorWithCode(err, ldaplib.LDAPResultNoSuchObject) {
			result.Passed = true
			result.Message = "Correctly rejected deletion of non-existent entry"
			logger.LogLDAPResult("Delete", "Delete", true, int(ldaplib.LDAPResultNoSuchObject), "No such object", duration)
			logger.Info("DeleteTest", "PASS: "+testName+" (rejected)", "duration", duration)
		} else {
			result.Passed = false
			result.Error = err
			result.Message = fmt.Sprintf("Failed with unexpected error: %v", err)
			logger.Error("DeleteTest", result.Message)
		}
	} else {
		result.Passed = false
		result.Message = "ERROR: Deletion of non-existent entry succeeded"
		logger.Error("DeleteTest", result.Message)
	}

	return result
}

// PerformCleanup deletes all tracked entries in reverse order
func PerformCleanup(conn *ldap.Connection, trk *tracker.Tracker) error {
	entries := trk.GetEntriesReversed()

	if len(entries) == 0 {
		logger.Info("Cleanup", "No entries to clean up")
		return nil
	}

	logger.Info("Cleanup", fmt.Sprintf("Starting cleanup of %d entries", len(entries)))

	successCount := 0
	failCount := 0

	for _, entry := range entries {
		logger.Debug("Cleanup", "Deleting entry", "dn", entry.DN, "type", entry.Type)

		delRequest := ldaplib.NewDelRequest(entry.DN, nil)
		err := conn.GetConnection().Del(delRequest)

		if err != nil {
			logger.Warn("Cleanup", "Failed to delete entry", "dn", entry.DN, "error", err)
			failCount++
		} else {
			logger.Info("Cleanup", "Successfully deleted entry", "dn", entry.DN)
			successCount++
		}
	}

	logger.Info("Cleanup", fmt.Sprintf("Cleanup complete: %d deleted, %d failed", successCount, failCount))

	if failCount > 0 {
		return fmt.Errorf("cleanup completed with %d failures", failCount)
	}

	return nil
}
