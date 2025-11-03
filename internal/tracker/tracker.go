package tracker

import (
	"fmt"
	"sync"
	"time"

	"ldap-automated-actions/internal/logger"
)

// EntryType represents the type of LDAP entry
type EntryType string

const (
	TypeOU    EntryType = "OU"
	TypeUser  EntryType = "User"
	TypeGroup EntryType = "Group"
	TypeOther EntryType = "Other"
)

// TrackedEntry represents a tracked LDAP entry
type TrackedEntry struct {
	DN        string
	Type      EntryType
	CreatedAt time.Time
}

// Tracker keeps track of all created LDAP entries for cleanup
type Tracker struct {
	entries []TrackedEntry
	mu      sync.Mutex
}

// NewTracker creates a new entry tracker
func NewTracker() *Tracker {
	return &Tracker{
		entries: make([]TrackedEntry, 0),
	}
}

// Track adds a new entry to the tracker
func (t *Tracker) Track(dn string, entryType EntryType) {
	t.mu.Lock()
	defer t.mu.Unlock()

	entry := TrackedEntry{
		DN:        dn,
		Type:      entryType,
		CreatedAt: time.Now(),
	}

	t.entries = append(t.entries, entry)
	logger.Debug("Tracker", "Tracking new entry", "dn", dn, "type", entryType)
}

// GetEntries returns all tracked entries
func (t *Tracker) GetEntries() []TrackedEntry {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Return a copy to prevent concurrent modification
	entries := make([]TrackedEntry, len(t.entries))
	copy(entries, t.entries)
	return entries
}

// GetEntriesReversed returns all tracked entries in reverse order (for cleanup)
func (t *Tracker) GetEntriesReversed() []TrackedEntry {
	t.mu.Lock()
	defer t.mu.Unlock()

	entries := make([]TrackedEntry, len(t.entries))
	for i, entry := range t.entries {
		entries[len(t.entries)-1-i] = entry
	}
	return entries
}

// Count returns the number of tracked entries
func (t *Tracker) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.entries)
}

// Clear removes all entries from the tracker
func (t *Tracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = make([]TrackedEntry, 0)
	logger.Debug("Tracker", "Cleared all tracked entries")
}

// PrintSummary prints a summary of all tracked entries
func (t *Tracker) PrintSummary() {
	entries := t.GetEntries()

	if len(entries) == 0 {
		fmt.Println("\nNo test data was created.")
		return
	}

	fmt.Printf("\n=== Created Test Data Summary ===\n")
	fmt.Printf("Total entries created: %d\n\n", len(entries))

	// Group by type
	byType := make(map[EntryType][]string)
	for _, entry := range entries {
		byType[entry.Type] = append(byType[entry.Type], entry.DN)
	}

	// Print by type
	for _, entryType := range []EntryType{TypeOU, TypeUser, TypeGroup, TypeOther} {
		if dns, ok := byType[entryType]; ok {
			fmt.Printf("%s entries (%d):\n", entryType, len(dns))
			for _, dn := range dns {
				fmt.Printf("  - %s\n", dn)
			}
			fmt.Println()
		}
	}

	fmt.Println("Note: Test data has been preserved. Use --cleanup flag to remove it automatically.")
}

// GetOldEntries returns entries older than the specified duration
func (t *Tracker) GetOldEntries(olderThan time.Duration) []TrackedEntry {
	t.mu.Lock()
	defer t.mu.Unlock()

	threshold := time.Now().Add(-olderThan)
	oldEntries := make([]TrackedEntry, 0)

	for _, entry := range t.entries {
		if entry.CreatedAt.Before(threshold) {
			oldEntries = append(oldEntries, entry)
		}
	}

	return oldEntries
}
