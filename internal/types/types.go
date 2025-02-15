package types

import "time"

// Journal represents a collection of entries
type Journal struct {
	Name     string    `json:"name"`
	Created  time.Time `json:"created"`
	EntryIDs []string  `json:"entry_ids"`
}

// Collection represents all journals and their metadata
type Collection struct {
	Journals       map[string]*Journal `json:"journals"`
	DefaultJournal string              `json:"default_journal"`
	NaClKeyID      string              `json:"nacl_key_id,omitempty"`
}

// Entry represents a single journal entry
type Entry struct {
	ID        string    `json:"id"`
	Created   time.Time `json:"created"`
	Body      []byte    `json:"body"`      // Encrypted content
	JournalID string    `json:"journalId"` // Reference to parent journal
}
