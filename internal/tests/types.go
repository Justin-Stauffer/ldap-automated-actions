package tests

import "time"

// TestResult represents the result of a single test
type TestResult struct {
	Name      string
	Operation string
	Passed    bool
	Duration  time.Duration
	Error     error
	Message   string
}

// TestSuite represents a collection of test results
type TestSuite struct {
	Name      string
	Results   []TestResult
	StartTime time.Time
	EndTime   time.Time
}

// GetStats returns statistics about the test suite
func (ts *TestSuite) GetStats() (total, passed, failed int, duration time.Duration) {
	total = len(ts.Results)
	for _, result := range ts.Results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
	}
	duration = ts.EndTime.Sub(ts.StartTime)
	return
}

// AllPassed returns true if all tests passed
func (ts *TestSuite) AllPassed() bool {
	for _, result := range ts.Results {
		if !result.Passed {
			return false
		}
	}
	return true
}
