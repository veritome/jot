package journal

import (
	"fmt"
	"time"

	"github.com/veritome/jot/internal/collection"
	"github.com/veritome/jot/internal/crypto"
	"github.com/veritome/jot/internal/entry"
	"github.com/veritome/jot/internal/types"
)

// Journal represents a collection of entries
type Journal struct {
	*types.Journal
}

// New creates a new journal with the given name
func New(name string) (*Journal, error) {
	// Verify NaCl keys exist
	keyPair, err := crypto.RestoreNaclFromBackup()
	if err != nil {
		return nil, fmt.Errorf("failed to restore NaCl keys: %w", err)
	}
	defer keyPair.Clear()

	return &Journal{
		Journal: &types.Journal{
			Name:     name,
			Created:  time.Now(),
			EntryIDs: make([]string, 0),
		},
	}, nil
}

// AsType converts the Journal to a types.Journal
func (j *Journal) AsType() *types.Journal {
	return j.Journal
}

// FromType creates a Journal from a types.Journal
func FromType(j *types.Journal) *Journal {
	return &Journal{Journal: j}
}

// Delete removes a journal and its associated data
func (j *Journal) Delete() error {
	// Load the collection to ensure we're working with the latest state
	coll, err := collection.Load()
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	// Remove all entries associated with this journal
	for _, entryID := range j.EntryIDs {
		entry, err := entry.Load(entryID)
		if err != nil {
			// Log error but continue with deletion
			fmt.Printf("Warning: Failed to load entry %s: %v\n", entryID, err)
			continue
		}
		if err := entry.Delete(); err != nil {
			fmt.Printf("Warning: Failed to delete entry %s: %v\n", entryID, err)
		}
	}

	// Remove the journal from the collection
	delete(coll.Journals, j.Name)

	// If this was the default journal, clear the default
	if coll.DefaultJournal == j.Name {
		coll.DefaultJournal = ""
	}

	// Save the updated collection
	if err := coll.Save(); err != nil {
		return fmt.Errorf("failed to save collection after journal deletion: %w", err)
	}

	return nil
}

// AddEntry adds a new entry to the journal
func (j *Journal) AddEntry(entryID string) error {
	j.EntryIDs = append(j.EntryIDs, entryID)

	// Load the collection to ensure we update the journal state
	coll, err := collection.Load()
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	// Update the journal in the collection
	coll.Journals[j.Name] = j.Journal

	// Save the updated collection
	if err := coll.Save(); err != nil {
		return fmt.Errorf("failed to save collection after adding entry: %w", err)
	}

	return nil
}

// GetEntries returns all entries in the journal
func (j *Journal) GetEntries() ([]*entry.Entry, error) {
	return entry.LoadJournalEntries(j.EntryIDs)
}

// Describe returns journal metadata
func (j *Journal) Describe() string {
	return fmt.Sprintf("Journal: %s\nCreated: %s\nEntries: %d",
		j.Name,
		j.Created.Format(time.RFC3339),
		len(j.EntryIDs))
}

// RemoveEntry removes an entry from the journal
func (j *Journal) RemoveEntry(entryID string) error {
	// Find and remove the entry ID from the journal's EntryIDs
	found := false
	newEntryIDs := make([]string, 0, len(j.EntryIDs))
	for _, id := range j.EntryIDs {
		if id == entryID {
			found = true
			continue
		}
		newEntryIDs = append(newEntryIDs, id)
	}

	if !found {
		return fmt.Errorf("entry %s not found in journal", entryID)
	}

	j.EntryIDs = newEntryIDs

	// Load the collection to ensure we update the journal state
	coll, err := collection.Load()
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	// Update the journal in the collection
	coll.Journals[j.Name] = j.Journal

	// Save the updated collection
	if err := coll.Save(); err != nil {
		return fmt.Errorf("failed to save collection after removing entry: %w", err)
	}

	return nil
}

// LoadAllJournals returns all journals from the collection
func LoadAllJournals() ([]*Journal, error) {
	coll, err := collection.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}

	journals := make([]*Journal, 0, len(coll.Journals))
	for _, j := range coll.Journals {
		journals = append(journals, &Journal{Journal: j})
	}

	return journals, nil
}
