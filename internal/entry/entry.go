package entry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/veritome/jot/internal/crypto"
	"github.com/veritome/jot/internal/types"
)

// Entry represents a single journal entry
type Entry struct {
	*types.Entry
}

// New creates a new entry with the given text
func New(journalID string, text string) (*Entry, error) {
	// Get NaCl keys
	keyPair, err := crypto.RestoreNaclFromBackup()
	if err != nil {
		return nil, fmt.Errorf("failed to restore NaCl keys: %w", err)
	}
	defer keyPair.Clear()

	// Encrypt the entry body
	encryptedBody, err := crypto.EncryptNacl(text, keyPair)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt entry with NaCl: %w", err)
	}

	return &Entry{
		Entry: &types.Entry{
			ID:        generateID(),
			Created:   time.Now(),
			Body:      encryptedBody,
			JournalID: journalID,
		},
	}, nil
}

// generateID creates a unique four-digit identifier for the entry
func generateID() string {
	// Get the entries directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Unable to access home directory")
	}

	entriesDir := filepath.Join(homeDir, ".jot", "entries")
	files, err := os.ReadDir(entriesDir)
	if err != nil {
		if os.IsNotExist(err) {
			// If the directory doesn't exist, start with 0001
			return "0001"
		}
		panic("Unable to read entries directory")
	}

	maxID := 0
	for _, file := range files {
		// Remove the .json extension and try to parse the ID
		name := strings.TrimSuffix(file.Name(), ".json")
		id, err := strconv.Atoi(name)
		if err == nil && id > maxID {
			maxID = id
		}
	}

	// Increment the maximum ID found and format as a four-digit string
	return fmt.Sprintf("%04d", maxID+1)
}

// GetDecryptedBody returns the decrypted entry content
func (e *Entry) GetDecryptedBody() (string, error) {
	keyPair, err := crypto.RestoreNaclFromBackup()
	if err != nil {
		return "", fmt.Errorf("failed to restore NaCl keys: %w", err)
	}
	defer keyPair.Clear()

	return crypto.DecryptNacl(e.Body, keyPair)
}

// Save persists the entry to storage
func (e *Entry) Save() error {
	data, err := json.MarshalIndent(e.Entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	entryPath, err := getEntryPath(e.ID)
	if err != nil {
		return fmt.Errorf("failed to get entry path: %w", err)
	}

	if err := os.WriteFile(entryPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write entry file: %w", err)
	}

	return nil
}

// Delete removes the entry from storage
func (e *Entry) Delete() error {
	entryPath, err := getEntryPath(e.ID)
	if err != nil {
		return fmt.Errorf("failed to get entry path: %w", err)
	}

	if err := os.Remove(entryPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete entry file: %w", err)
	}

	return nil
}

// Load loads an entry from storage by its ID
func Load(id string) (*Entry, error) {
	entryPath, err := getEntryPath(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry path: %w", err)
	}

	data, err := os.ReadFile(entryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read entry file: %w", err)
	}

	var entry types.Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal entry: %w", err)
	}

	return &Entry{Entry: &entry}, nil
}

// getEntryPath returns the path where an entry should be stored
func getEntryPath(id string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	entriesDir := filepath.Join(homeDir, ".jot", "entries")
	if err := os.MkdirAll(entriesDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create entries directory: %w", err)
	}

	return filepath.Join(entriesDir, fmt.Sprintf("%s.json", id)), nil
}

// LoadJournalEntries loads all entries for a given journal
func LoadJournalEntries(entryIDs []string) ([]*Entry, error) {
	entries := make([]*Entry, 0, len(entryIDs))
	for _, id := range entryIDs {
		e, err := Load(id)
		if err != nil {
			return nil, fmt.Errorf("failed to load entry %s: %w", id, err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}
