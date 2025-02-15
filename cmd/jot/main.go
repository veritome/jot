package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/veritome/jot/internal/collection"
	"github.com/veritome/jot/internal/crypto"
	"github.com/veritome/jot/internal/entry"
	"github.com/veritome/jot/internal/journal"
)

var journalCollection *collection.Collection

var collectionCommands = map[string]bool{
	"collection": true,
	"c":          true,
}

var journalCommands = map[string]bool{
	"journal": true,
	"j":       true,
}

func init() {
	var err error
	journalCollection, err = collection.Load()
	if err != nil {
		fmt.Printf("Error loading collection: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	journalFlag := flag.String("journal", "", "Specify journal name for the entry")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		const usage = `Usage: jot [OPTIONS] [COMMAND] [ARGS...]
    __
   |  |
   |  |_____ _____ 
 __|  |     |_   _|
|  |  |  |  | | |  
|  |  |_____| |_|  
|_____|

A simple, secure journaling tool.

Options:
  -j, --journal <name>    Specify journal name for the entry

Commands:
  <entry text>            Create a new entry in the default journal
  collection, c           List all journals
  journal, j <command>    Manage journals
  nuke                    Delete all data and reset JOT

Journal Commands:
  new <name>             Create a new journal
  delete <name>          Delete an existing journal
  default <name>         Set the default journal
  read <name>            Display all entries in a journal
  describe <name>        Show journal metadata
  delete-entry <name> <id>  Delete an entry from a journal

Examples:
  jot "Had a great day today"                    Create entry in default journal
  jot -j work "Important meeting notes"          Create entry in "work" journal
  jot journal new work                           Create a new journal called "work"
  jot journal read work                          Read all entries in "work" journal
  jot journal delete-entry work 0001             Delete entry 0001 from "work" journal

For more information, visit: https://github.com/veritome/jot`
		fmt.Println(usage)
		os.Exit(1)
	}

	// Handle nuke command
	if args[0] == "nuke" {
		handleNukeCommand()
		return
	}

	// Handle collection command
	if collectionCommands[args[0]] {
		if len(args) != 1 {
			fmt.Println("Usage: jot collection")
			os.Exit(1)
		}
		handleCollectionCommand()
		return
	}

	// Handle journal management commands
	if journalCommands[args[0]] {
		if len(args) < 2 {
			fmt.Println("Usage: jot journal <new|delete|default|read|describe|delete-entry> [args]")
			os.Exit(1)
		}
		handleJournalCommand(args[1:])
		return
	}

	// At this point, all remaining args should be considered entry text
	// No need to process args[0] differently as it's not a command
	entryText := strings.Join(args, " ")
	handleEntry(*journalFlag, entryText)
}

func handleCollectionCommand() {
	journals := journalCollection.List()
	if len(journals) == 0 {
		fmt.Println("No journals found")
		return
	}

	// Sort journals alphabetically
	sort.Strings(journals)

	fmt.Println("Available Journals:")
	fmt.Println("------------------")
	for _, name := range journals {
		fmt.Printf("  %s\n", name)
	}
	fmt.Println("\nNote: * indicates default journal")
}

func handleJournalCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Missing journal command")
		os.Exit(1)
	}

	switch args[0] {
	case "new":
		if len(args) != 2 {
			fmt.Println("Usage: jot journal new <name>")
			os.Exit(1)
		}
		j, err := journal.New(args[1])
		if err != nil {
			fmt.Printf("Error creating journal: %v\n", err)
			os.Exit(1)
		}
		if err := journalCollection.AddJournal(j.AsType()); err != nil {
			fmt.Printf("Error adding journal: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Created journal: %s\n", args[1])

	case "delete":
		if len(args) != 2 {
			fmt.Println("Usage: jot journal delete <name>")
			os.Exit(1)
		}
		if err := journalCollection.RemoveJournal(args[1]); err != nil {
			fmt.Printf("Error deleting journal: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted journal: %s\n", args[1])

	case "default":
		if len(args) != 2 {
			fmt.Println("Usage: jot journal default <name>")
			os.Exit(1)
		}
		if err := journalCollection.SetDefaultJournal(args[1]); err != nil {
			fmt.Printf("Error setting default journal: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set default journal to: %s\n", args[1])

	case "read":
		if len(args) != 2 {
			fmt.Println("Usage: jot journal read <name>")
			os.Exit(1)
		}
		j, exists := journalCollection.Journals[args[1]]
		if !exists {
			fmt.Printf("Journal '%s' does not exist\n", args[1])
			os.Exit(1)
		}

		wrappedJ := journal.FromType(j)
		entries, err := wrappedJ.GetEntries()
		if err != nil {
			fmt.Printf("Error reading entries: %v\n", err)
			os.Exit(1)
		}

		if len(entries) == 0 {
			fmt.Printf("No entries found in journal '%s'\n", args[1])
			return
		}

		fmt.Printf("Entries in journal '%s':\n", args[1])
		fmt.Println("------------------------")
		for _, e := range entries {
			content, err := e.GetDecryptedBody()
			if err != nil {
				fmt.Printf("Error decrypting entry %s: %v\n", e.ID, err)
				continue
			}
			fmt.Printf("[%s] %s\n", e.Created.Format("2006-01-02 15:04:05"), content)
		}

	case "describe":
		if len(args) != 2 {
			fmt.Println("Usage: jot journal describe <name>")
			os.Exit(1)
		}
		j, exists := journalCollection.Journals[args[1]]
		if !exists {
			fmt.Printf("Journal '%s' does not exist\n", args[1])
			os.Exit(1)
		}

		wrappedJ := journal.FromType(j)
		fmt.Println(wrappedJ.Describe())

	case "delete-entry":
		if len(args) != 3 {
			fmt.Println("Usage: jot journal delete-entry <journal-name> <entry-id>")
			os.Exit(1)
		}
		journalName := args[1]
		entryID := args[2]

		j, exists := journalCollection.Journals[journalName]
		if !exists {
			fmt.Printf("Journal '%s' does not exist\n", journalName)
			os.Exit(1)
		}

		wrappedJ := journal.FromType(j)
		// Load the entry to verify it exists
		e, err := entry.Load(entryID)
		if err != nil {
			fmt.Printf("Error loading entry: %v\n", err)
			os.Exit(1)
		}

		// Verify the entry belongs to the specified journal
		if e.JournalID != journalName {
			fmt.Printf("Entry %s does not belong to journal '%s'\n", entryID, journalName)
			os.Exit(1)
		}

		// Delete the entry from storage
		if err := e.Delete(); err != nil {
			fmt.Printf("Error deleting entry: %v\n", err)
			os.Exit(1)
		}

		// Remove the entry from the journal
		if err := wrappedJ.RemoveEntry(entryID); err != nil {
			fmt.Printf("Error removing entry from journal: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Entry %s deleted from journal '%s'\n", entryID, journalName)

	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		os.Exit(1)
	}
}

func handleEntry(journalName, text string) {
	if journalName == "" {
		journalName = journalCollection.GetDefaultJournal()
		if journalName == "" {
			fmt.Println("No default journal set. Please specify a journal with --journal or set a default journal.")
			os.Exit(1)
		}
	}

	// Get the journal
	j, exists := journalCollection.Journals[journalName]
	if !exists {
		fmt.Printf("Journal '%s' does not exist\n", journalName)
		os.Exit(1)
	}

	wrappedJ := journal.FromType(j)

	// Create new entry
	e, err := entry.New(journalName, text)
	if err != nil {
		fmt.Printf("Error creating entry: %v\n", err)
		os.Exit(1)
	}

	// Save the entry
	if err := e.Save(); err != nil {
		fmt.Printf("Error saving entry: %v\n", err)
		os.Exit(1)
	}

	// Add entry to journal
	if err := wrappedJ.AddEntry(e.ID); err != nil {
		fmt.Printf("Error adding entry to journal: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Entry added to journal '%s'\n", journalName)
}

func handleNukeCommand() {
	fmt.Print("WARNING: This will delete all journals and entries. Are you sure? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	response = strings.TrimSpace(response)
	if response != "y" && response != "Y" {
		fmt.Println("Operation cancelled")
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	// Remove .jot directory
	jotDir := filepath.Join(homeDir, ".jot")
	if err := os.RemoveAll(jotDir); err != nil {
		fmt.Printf("Error removing .jot directory: %v\n", err)
		os.Exit(1)
	}

	// Generate new NaCl keys
	if _, err := crypto.GenerateNaclKey(); err != nil {
		fmt.Printf("Error generating new NaCl keys: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All data has been deleted and encryption keys have been regenerated.")
}
