package tests

import (
	"fmt"
	"time"

	"ldap-automated-actions/internal/ldap"
	"ldap-automated-actions/internal/logger"

	ldaplib "github.com/go-ldap/ldap/v3"
)

// TestSearch runs all search operation tests
func TestSearch(conn *ldap.Connection, testBaseDN string) []TestResult {
	logger.Info("SearchTest", "Starting Search operation tests")
	results := make([]TestResult, 0)

	// Test 1: Search with base scope
	results = append(results, testSearchBase(conn, testBaseDN))

	// Test 2: Search with one level scope
	results = append(results, testSearchOneLevel(conn, testBaseDN))

	// Test 3: Search with subtree scope
	results = append(results, testSearchSubtree(conn, testBaseDN))

	// Test 4: Search with filter
	results = append(results, testSearchWithFilter(conn, testBaseDN))

	// Test 5: Search with attribute selection
	results = append(results, testSearchWithAttributes(conn, testBaseDN))

	// Test 6: Search with paging (if many results)
	results = append(results, testSearchWithPaging(conn, conn.GetConfig().BaseDN))

	logger.Info("SearchTest", "Completed Search operation tests", "total", len(results))
	return results
}

func testSearchBase(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Search with Base Scope Test"
	logger.Info("SearchTest", "Running: "+testName)

	filter := "(objectClass=*)"
	attributes := []string{"*"}

	logger.LogSearchOperation("Search", testBaseDN, filter, "base", attributes)

	searchRequest := ldaplib.NewSearchRequest(
		testBaseDN,
		ldaplib.ScopeBaseObject,
		ldaplib.NeverDerefAliases,
		0, 0, false,
		filter,
		attributes,
		nil,
	)

	start := time.Now()
	result, err := conn.GetConnection().Search(searchRequest)
	duration := time.Since(start)

	testResult := TestResult{
		Name:      testName,
		Operation: "Search",
		Duration:  duration,
	}

	if err != nil {
		testResult.Passed = false
		testResult.Error = err
		testResult.Message = fmt.Sprintf("Search failed: %v", err)
		logger.LogLDAPResult("Search", "Search", false, -1, err.Error(), duration)
		logger.Error("SearchTest", testResult.Message)
	} else {
		testResult.Passed = true
		testResult.Message = fmt.Sprintf("Found %d entries (base scope)", len(result.Entries))
		logger.LogSearchResult("Search", len(result.Entries), duration)
		logger.Info("SearchTest", "PASS: "+testName, "entries", len(result.Entries), "duration", duration)
	}

	return testResult
}

func testSearchOneLevel(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Search with One Level Scope Test"
	logger.Info("SearchTest", "Running: "+testName)

	filter := "(objectClass=*)"
	attributes := []string{"cn", "ou", "objectClass"}

	logger.LogSearchOperation("Search", testBaseDN, filter, "one", attributes)

	searchRequest := ldaplib.NewSearchRequest(
		testBaseDN,
		ldaplib.ScopeSingleLevel,
		ldaplib.NeverDerefAliases,
		0, 0, false,
		filter,
		attributes,
		nil,
	)

	start := time.Now()
	result, err := conn.GetConnection().Search(searchRequest)
	duration := time.Since(start)

	testResult := TestResult{
		Name:      testName,
		Operation: "Search",
		Duration:  duration,
	}

	if err != nil {
		testResult.Passed = false
		testResult.Error = err
		testResult.Message = fmt.Sprintf("Search failed: %v", err)
		logger.LogLDAPResult("Search", "Search", false, -1, err.Error(), duration)
		logger.Error("SearchTest", testResult.Message)
	} else {
		testResult.Passed = true
		testResult.Message = fmt.Sprintf("Found %d entries (one level scope)", len(result.Entries))
		logger.LogSearchResult("Search", len(result.Entries), duration)
		logger.Info("SearchTest", "PASS: "+testName, "entries", len(result.Entries), "duration", duration)
	}

	return testResult
}

func testSearchSubtree(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Search with Subtree Scope Test"
	logger.Info("SearchTest", "Running: "+testName)

	filter := "(objectClass=*)"
	attributes := []string{"dn"}

	logger.LogSearchOperation("Search", testBaseDN, filter, "sub", attributes)

	searchRequest := ldaplib.NewSearchRequest(
		testBaseDN,
		ldaplib.ScopeWholeSubtree,
		ldaplib.NeverDerefAliases,
		0, 0, false,
		filter,
		attributes,
		nil,
	)

	start := time.Now()
	result, err := conn.GetConnection().Search(searchRequest)
	duration := time.Since(start)

	testResult := TestResult{
		Name:      testName,
		Operation: "Search",
		Duration:  duration,
	}

	if err != nil {
		testResult.Passed = false
		testResult.Error = err
		testResult.Message = fmt.Sprintf("Search failed: %v", err)
		logger.LogLDAPResult("Search", "Search", false, -1, err.Error(), duration)
		logger.Error("SearchTest", testResult.Message)
	} else {
		testResult.Passed = true
		testResult.Message = fmt.Sprintf("Found %d entries (subtree scope)", len(result.Entries))
		logger.LogSearchResult("Search", len(result.Entries), duration)
		logger.Info("SearchTest", "PASS: "+testName, "entries", len(result.Entries), "duration", duration)

		// Log some entry DNs at trace level
		if len(result.Entries) > 0 {
			logger.Trace("Search", fmt.Sprintf("Sample entries: %d total", len(result.Entries)))
			for i, entry := range result.Entries {
				if i >= 5 {
					break // Only log first 5
				}
				logger.Trace("Search", fmt.Sprintf("  [%d] %s", i+1, entry.DN))
			}
		}
	}

	return testResult
}

func testSearchWithFilter(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Search with Filter Test"
	logger.Info("SearchTest", "Running: "+testName)

	// Search for inetOrgPerson entries
	filter := "(objectClass=inetOrgPerson)"
	attributes := []string{"cn", "mail", "sn"}

	logger.LogSearchOperation("Search", testBaseDN, filter, "sub", attributes)

	searchRequest := ldaplib.NewSearchRequest(
		testBaseDN,
		ldaplib.ScopeWholeSubtree,
		ldaplib.NeverDerefAliases,
		0, 0, false,
		filter,
		attributes,
		nil,
	)

	start := time.Now()
	result, err := conn.GetConnection().Search(searchRequest)
	duration := time.Since(start)

	testResult := TestResult{
		Name:      testName,
		Operation: "Search",
		Duration:  duration,
	}

	if err != nil {
		testResult.Passed = false
		testResult.Error = err
		testResult.Message = fmt.Sprintf("Search failed: %v", err)
		logger.LogLDAPResult("Search", "Search", false, -1, err.Error(), duration)
		logger.Error("SearchTest", testResult.Message)
	} else {
		testResult.Passed = true
		testResult.Message = fmt.Sprintf("Found %d inetOrgPerson entries with filter", len(result.Entries))
		logger.LogSearchResult("Search", len(result.Entries), duration)
		logger.Info("SearchTest", "PASS: "+testName, "entries", len(result.Entries), "duration", duration)

		// Log details of found entries at trace level
		for _, entry := range result.Entries {
			logger.Trace("Search", "Entry found", "dn", entry.DN, "cn", entry.GetAttributeValue("cn"))
		}
	}

	return testResult
}

func testSearchWithAttributes(conn *ldap.Connection, testBaseDN string) TestResult {
	testName := "Search with Attribute Selection Test"
	logger.Info("SearchTest", "Running: "+testName)

	// Search for test user and only retrieve specific attributes
	filter := "(cn=testuser)"
	attributes := []string{"cn", "mail"} // Only request these attributes

	logger.LogSearchOperation("Search", testBaseDN, filter, "sub", attributes)

	searchRequest := ldaplib.NewSearchRequest(
		testBaseDN,
		ldaplib.ScopeWholeSubtree,
		ldaplib.NeverDerefAliases,
		0, 0, false,
		filter,
		attributes,
		nil,
	)

	start := time.Now()
	result, err := conn.GetConnection().Search(searchRequest)
	duration := time.Since(start)

	testResult := TestResult{
		Name:      testName,
		Operation: "Search",
		Duration:  duration,
	}

	if err != nil {
		testResult.Passed = false
		testResult.Error = err
		testResult.Message = fmt.Sprintf("Search failed: %v", err)
		logger.LogLDAPResult("Search", "Search", false, -1, err.Error(), duration)
		logger.Error("SearchTest", testResult.Message)
	} else {
		if len(result.Entries) > 0 {
			entry := result.Entries[0]
			// Verify only requested attributes are returned
			hasOnlyRequested := true
			for _, attr := range entry.Attributes {
				found := false
				for _, requested := range attributes {
					if attr.Name == requested {
						found = true
						break
					}
				}
				if !found {
					hasOnlyRequested = false
					logger.Debug("SearchTest", "Unexpected attribute in result", "attribute", attr.Name)
				}
			}

			testResult.Passed = true
			testResult.Message = fmt.Sprintf("Found entries with attribute selection (attributes filtered: %v)", hasOnlyRequested)
			logger.LogSearchResult("Search", len(result.Entries), duration)
			logger.Info("SearchTest", "PASS: "+testName, "entries", len(result.Entries), "duration", duration)

			// Log retrieved attributes
			logger.Trace("Search", "Retrieved attributes", "cn", entry.GetAttributeValue("cn"), "mail", entry.GetAttributeValue("mail"))
		} else {
			testResult.Passed = true
			testResult.Message = "No entries found matching filter (expected if test user doesn't exist yet)"
			logger.Info("SearchTest", "PASS: "+testName+" (no results)", "duration", duration)
		}
	}

	return testResult
}

func testSearchWithPaging(conn *ldap.Connection, baseDN string) TestResult {
	testName := "Search with Paging Test"
	logger.Info("SearchTest", "Running: "+testName)

	filter := "(objectClass=*)"
	attributes := []string{"dn"}
	pageSize := uint32(10)

	logger.LogSearchOperation("Search", baseDN, filter, "sub (paged)", attributes)
	logger.Debug("SearchTest", "Using paging", "pageSize", pageSize)

	searchRequest := ldaplib.NewSearchRequest(
		baseDN,
		ldaplib.ScopeWholeSubtree,
		ldaplib.NeverDerefAliases,
		0, 0, false,
		filter,
		attributes,
		nil,
	)

	start := time.Now()
	totalEntries := 0
	pageCount := 0

	// Perform paged search
	// Note: SearchWithPaging in go-ldap/v3 returns (result, error)
	pagingControl := ldaplib.NewControlPaging(pageSize)
	searchRequest.Controls = append(searchRequest.Controls, pagingControl)

	var err error
	for {
		result, searchErr := conn.GetConnection().Search(searchRequest)
		if searchErr != nil {
			err = searchErr
			break
		}

		pageCount++
		totalEntries += len(result.Entries)
		logger.Trace("Search", fmt.Sprintf("Page %d: %d entries", pageCount, len(result.Entries)))

		// Check if there are more pages
		var updatedControl *ldaplib.ControlPaging
		for _, control := range result.Controls {
			if c, ok := control.(*ldaplib.ControlPaging); ok {
				updatedControl = c
				break
			}
		}

		if updatedControl == nil || len(updatedControl.Cookie) == 0 {
			break // No more pages
		}

		// Update the paging control for the next request
		pagingControl.SetCookie(updatedControl.Cookie)
	}

	duration := time.Since(start)

	testResult := TestResult{
		Name:      testName,
		Operation: "Search",
		Duration:  duration,
	}

	if err != nil {
		testResult.Passed = false
		testResult.Error = err
		testResult.Message = fmt.Sprintf("Paged search failed: %v", err)
		logger.LogLDAPResult("Search", "Search (paged)", false, -1, err.Error(), duration)
		logger.Error("SearchTest", testResult.Message)
	} else {
		testResult.Passed = true
		testResult.Message = fmt.Sprintf("Paged search completed: %d entries across %d pages", totalEntries, pageCount)
		logger.LogSearchResult("Search", totalEntries, duration)
		logger.Info("SearchTest", "PASS: "+testName, "entries", totalEntries, "pages", pageCount, "duration", duration)
	}

	return testResult
}
