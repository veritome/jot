package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/veritome/jot/internal/entry"
	"github.com/veritome/jot/internal/journal"
)

// Common styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("205"))

	paginationStyle = list.DefaultStyles().PaginationStyle.
			PaddingLeft(2)

	helpStyle = list.DefaultStyles().HelpStyle.
			PaddingLeft(2).
			PaddingBottom(1)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Bright red
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("196")).
			Padding(1, 3).
			Margin(1, 0)

	confirmationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).   // Black text
				Background(lipgloss.Color("226")). // Yellow background
				Bold(true).
				PaddingLeft(2).
				PaddingRight(2).
				MarginTop(1)
)

// Package ui provides the terminal user interface components for the jot application.
// It implements list views and interactive menus for managing journal entries.

// entryItem represents a journal entry in the list view.
// It implements the list.Item interface from charmbracelet/bubbles.
type entryItem struct {
	id           string // Unique identifier for the entry
	content      string // Decrypted content of the entry
	created      string // Creation timestamp
	marked       bool   // Whether the entry is marked for deletion
	isDeleteList bool   // Whether this item is in a deletion list view
}

func (i entryItem) Title() string {
	if i.isDeleteList {
		mark := " "
		if i.marked {
			mark = "X"
		}
		return fmt.Sprintf("[%s] %s", mark, i.id)
	}
	return i.id
}

func (i entryItem) Description() string {
	return fmt.Sprintf("%s | %s", i.created, i.content)
}

func (i entryItem) FilterValue() string {
	return i.content
}

// ListEntriesModel represents the view model for displaying journal entries.
// It provides a scrollable list interface for viewing entries.
type ListEntriesModel struct {
	list     list.Model       // The underlying list UI component
	journal  *journal.Journal // Reference to the journal being displayed
	quitting bool             // Whether the view is being closed
}

// NewListEntriesModel creates a new model for listing entries
func NewListEntriesModel(j *journal.Journal) (*ListEntriesModel, error) {
	entries, err := j.GetEntries()
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	items := make([]list.Item, 0, len(entries))
	for _, e := range entries {
		content, err := e.GetDecryptedBody()
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt entry %s: %w", e.ID, err)
		}
		items = append(items, entryItem{
			id:           e.ID,
			content:      content,
			created:      e.Created.Format(time.RFC3339),
			isDeleteList: false,
		})
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.SelectedDesc = selectedItemStyle

	// Calculate height: 2 lines per item (title + description) + 3 for title and borders
	height := len(items)*2 + 3

	l := list.New(items, delegate, 0, height)
	l.Title = "Journal Entries"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return &ListEntriesModel{
		list:    l,
		journal: j,
	}, nil
}

func (m ListEntriesModel) Init() tea.Cmd {
	return nil
}

func (m ListEntriesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))):
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := itemStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ListEntriesModel) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}

// DeleteEntriesModel represents the view model for the deletion interface.
// It provides a multi-select interface for choosing entries to delete.
type DeleteEntriesModel struct {
	list          list.Model       // The underlying list UI component
	journal       *journal.Journal // Reference to the journal being modified
	quitting      bool             // Whether the view is being closed
	items         []entryItem      // List of entries that can be deleted
	keys          keyMap           // Key bindings for the delete interface
	confirmDelete bool             // Whether deletion has been confirmed
	markedCount   int              // Number of entries marked for deletion
	deleteResult  *DeleteResult    // Result of the deletion operation
}

// keyMap defines the key bindings for the delete interface
type keyMap struct {
	space key.Binding
	enter key.Binding
	quit  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.space, k.enter, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.space, k.enter, k.quit},
	}
}

// NewDeleteEntriesModel creates a new model for deleting entries.
// It loads all entries from the journal and prepares them for potential deletion.
func NewDeleteEntriesModel(j *journal.Journal) (*DeleteEntriesModel, error) {
	entries, err := j.GetEntries()
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	items := make([]entryItem, 0, len(entries))
	listItems := make([]list.Item, 0, len(entries))
	for _, e := range entries {
		content, err := e.GetDecryptedBody()
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt entry %s: %w", e.ID, err)
		}
		item := entryItem{
			id:           e.ID,
			content:      content,
			created:      e.Created.Format(time.RFC3339),
			marked:       false,
			isDeleteList: true,
		}
		items = append(items, item)
		listItems = append(listItems, item)
	}

	keys := keyMap{
		space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle selection"),
		),
		enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm selection"),
		),
		quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "quit"),
		),
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.SelectedDesc = selectedItemStyle

	// Calculate height: 2 lines per item (title + description) + 3 for title and borders + 1 for help text
	height := len(items)*2 + 4

	l := list.New(listItems, delegate, 0, height)
	l.Title = "Select Entries to Delete (Space to select, Enter to confirm)"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.SetShowHelp(true)
	l.AdditionalShortHelpKeys = keys.ShortHelp
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)

	return &DeleteEntriesModel{
		list:          l,
		journal:       j,
		items:         items,
		keys:          keys,
		confirmDelete: false,
		markedCount:   0,
	}, nil
}

func (m *DeleteEntriesModel) Init() tea.Cmd {
	return nil
}

// Message type for deletion completion
type entriesDeletedMsg struct {
	count   int
	journal string
}

// DeleteResult contains information about the deletion operation
type DeleteResult struct {
	Count   int
	Journal string
}

func (m *DeleteEntriesModel) deleteEntries(ids []string) tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			for _, id := range ids {
				e, err := entry.Load(id)
				if err != nil {
					fmt.Printf("Error loading entry %s: %v\n", id, err)
					continue
				}

				if err := e.Delete(); err != nil {
					fmt.Printf("Error deleting entry %s: %v\n", id, err)
					continue
				}

				if err := m.journal.RemoveEntry(id); err != nil {
					fmt.Printf("Error removing entry %s from journal: %v\n", id, err)
				}
			}

			// Remove deleted items from the list
			var newItems []entryItem
			for _, item := range m.items {
				if !contains(ids, item.id) {
					newItems = append(newItems, item)
				}
			}
			m.items = newItems
			listItems := make([]list.Item, len(m.items))
			for i, item := range m.items {
				listItems[i] = item
			}
			m.list.SetItems(listItems)

			return entriesDeletedMsg{
				count:   len(ids),
				journal: m.journal.Name,
			}
		},
		tea.Quit,
	)
}

func (m *DeleteEntriesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.list.SettingFilter() {
			switch {
			case key.Matches(msg, m.keys.quit):
				m.quitting = true
				return m, tea.Quit
			case key.Matches(msg, m.keys.space):
				// Only allow marking if not in confirmation mode
				if !m.confirmDelete {
					index := m.list.Index()
					if index < len(m.items) {
						m.items[index].marked = !m.items[index].marked
						// Update marked count
						if m.items[index].marked {
							m.markedCount++
						} else {
							m.markedCount--
						}
						listItems := make([]list.Item, len(m.items))
						for i, item := range m.items {
							listItems[i] = item
						}
						m.list.SetItems(listItems)
					}
				}
				return m, nil
			case key.Matches(msg, m.keys.enter):
				if !m.confirmDelete {
					// First enter press shows confirmation
					if m.markedCount > 0 {
						m.confirmDelete = true
						m.list.Title = fmt.Sprintf("Are you sure you want to delete %d entries? (Enter to confirm, Esc to cancel)", m.markedCount)
						return m, nil
					}
				} else {
					// Second enter press performs deletion
					var markedIDs []string
					for _, item := range m.items {
						if item.marked {
							markedIDs = append(markedIDs, item.id)
						}
					}
					return m, m.deleteEntries(markedIDs)
				}
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		h, v := itemStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case entriesDeletedMsg:
		m.quitting = true
		m.deleteResult = &DeleteResult{
			Count:   msg.count,
			Journal: msg.journal,
		}
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *DeleteEntriesModel) View() string {
	if m.quitting {
		return ""
	}

	view := m.list.View()

	// Add confirmation message if in confirmation mode
	if m.confirmDelete {
		warning := warningStyle.Render("⚠️  WARNING: This action cannot be undone!")
		entryText := "entries"
		if m.markedCount == 1 {
			entryText = "entry"
		}
		confirmation := confirmationStyle.Render(fmt.Sprintf("Press ENTER to delete %d %s or ESC to cancel", m.markedCount, entryText))
		view = fmt.Sprintf("%s\n\n%s\n%s", view, warning, confirmation)
	}

	return view
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// HandleShowEntries displays entries in a journal
func HandleShowEntries(j *journal.Journal) error {
	model, err := NewListEntriesModel(j)
	if err != nil {
		return fmt.Errorf("failed to create list model: %w", err)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run program: %w", err)
	}

	return nil
}

// HandleInteractiveDelete handles interactive deletion of entries
func HandleInteractiveDelete(j *journal.Journal) error {
	model, err := NewDeleteEntriesModel(j)
	if err != nil {
		return fmt.Errorf("failed to create delete model: %w", err)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run program: %w", err)
	}

	// After returning to normal screen, print deletion result if any
	if deleteModel, ok := m.(*DeleteEntriesModel); ok && deleteModel.deleteResult != nil {
		if deleteModel.deleteResult.Count == 1 {
			fmt.Printf("%d journal entry was deleted from %s\n", deleteModel.deleteResult.Count, deleteModel.deleteResult.Journal)
		} else {
			fmt.Printf("%d journal entries were deleted from %s\n", deleteModel.deleteResult.Count, deleteModel.deleteResult.Journal)
		}
	}

	return nil
}
