package collection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/veritome/jot/internal/crypto"
	"github.com/veritome/jot/internal/types"
)

// Collection represents all journals and their metadata
type Collection struct {
	*types.Collection
}

// NewCollection creates a new journal collection
func NewCollection() (*Collection, error) {
	return &Collection{
		Collection: &types.Collection{
			Journals: make(map[string]*types.Journal),
		},
	}, nil
}

// Save persists the collection to disk
func (c *Collection) Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	jotDir := filepath.Join(homeDir, ".jot")
	if err := os.MkdirAll(jotDir, 0700); err != nil {
		return fmt.Errorf("failed to create jot directory: %w", err)
	}

	data, err := json.MarshalIndent(c.Collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal collection: %w", err)
	}

	if err := os.WriteFile(filepath.Join(jotDir, "collection.json"), data, 0600); err != nil {
		return fmt.Errorf("failed to write collection file: %w", err)
	}

	return nil
}

// Load loads the collection from disk
func Load() (*Collection, error) {
	// Try to restore NaCl keys first
	keyPair, err := crypto.RestoreNaclFromBackup()
	if err != nil {
		// If keys don't exist, generate them
		if _, err := crypto.GenerateNaclKey(); err != nil {
			return nil, fmt.Errorf("failed to generate NaCl keys: %w", err)
		}
	} else {
		keyPair.Clear() // Clear the keys from memory
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	collectionPath := filepath.Join(homeDir, ".jot", "collection.json")
	if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
		// If collection doesn't exist, create a new one
		return NewCollection()
	}

	data, err := os.ReadFile(collectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read collection file: %w", err)
	}

	var collection types.Collection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("failed to unmarshal collection: %w", err)
	}

	return &Collection{Collection: &collection}, nil
}

// SetDefaultJournal sets the specified journal as the default
func (c *Collection) SetDefaultJournal(name string) error {
	if _, exists := c.Journals[name]; !exists {
		return fmt.Errorf("journal '%s' does not exist", name)
	}
	c.DefaultJournal = name
	return c.Save()
}

// GetDefaultJournal returns the name of the default journal
func (c *Collection) GetDefaultJournal() string {
	return c.DefaultJournal
}

// List returns a formatted list of all journals, with an asterisk next to the default
func (c *Collection) List() []string {
	var list []string
	for name := range c.Journals {
		if name == c.DefaultJournal {
			list = append(list, fmt.Sprintf("%s *", name))
		} else {
			list = append(list, name)
		}
	}
	return list
}

// AddJournal adds a journal to the collection and sets it as default if it's the first one
func (c *Collection) AddJournal(j *types.Journal) error {
	if _, exists := c.Journals[j.Name]; exists {
		return fmt.Errorf("journal '%s' already exists", j.Name)
	}

	c.Journals[j.Name] = j

	// If this is the first journal, set it as default
	if len(c.Journals) == 1 {
		c.DefaultJournal = j.Name
	}

	return c.Save()
}

// RemoveJournal removes a journal from the collection
func (c *Collection) RemoveJournal(name string) error {
	if _, exists := c.Journals[name]; !exists {
		return fmt.Errorf("journal '%s' does not exist", name)
	}
	if name == c.DefaultJournal {
		c.DefaultJournal = ""
	}
	delete(c.Journals, name)
	return c.Save()
}

