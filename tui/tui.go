package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/luca/jotta-archiver/archive"
	"github.com/luca/jotta-archiver/config"
)

// Screen represents the current screen state
type Screen int

const (
	ScreenPresetSelection Screen = iota
	ScreenCustomRemote
	ScreenNameEditing
)

// Model represents the TUI state
type Model struct {
	screen        Screen
	presets       []config.Preset
	selectedIndex int
	archiveName   string
	textInput     textinput.Model
	remoteInput   textinput.Model
	folderPath    string
	customRemote  string
	uploadID      string
	ctx           context.Context
	cancel        context.CancelFunc
	err           error
	quitting      bool
	debugMode     bool
}

// New creates a new TUI model
func New(presets []config.Preset, initialName, folderPath string) Model {
	ti := textinput.New()
	ti.Placeholder = "Archive name"
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50
	ti.SetValue(initialName)

	ri := textinput.New()
	ri.Placeholder = "/path/to/remote/directory"
	ri.CharLimit = 200
	ri.Width = 50

	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		screen:        ScreenPresetSelection,
		presets:       presets,
		selectedIndex: 0,
		archiveName:   initialName,
		textInput:     ti,
		remoteInput:   ri,
		folderPath:    folderPath,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case ScreenPresetSelection:
			return m.updatePresetSelection(msg)
		case ScreenCustomRemote:
			return m.updateCustomRemote(msg)
		case ScreenNameEditing:
			return m.updateNameEditing(msg)
		}

	case tea.WindowSizeMsg:
		return m, nil
	}

	return m, nil
}

func (m Model) updatePresetSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		m.cancel()
		return m, tea.Quit

	case "d":
		m.debugMode = !m.debugMode

	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case "down", "j":
		// +1 for the "Custom..." option
		if m.selectedIndex < len(m.presets) {
			m.selectedIndex++
		}

	case "enter":
		// Check if custom option is selected (last item)
		if m.selectedIndex == len(m.presets) {
			m.remoteInput.Focus()
			m.screen = ScreenCustomRemote
			return m, nil
		}
		m.screen = ScreenNameEditing
		return m, nil
	}

	return m, nil
}

func (m Model) updateCustomRemote(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		m.cancel()
		return m, tea.Quit

	case "esc":
		m.screen = ScreenPresetSelection
		return m, nil

	case "enter":
		m.customRemote = m.remoteInput.Value()
		if m.customRemote == "" {
			return m, nil // Don't proceed with empty remote
		}
		m.textInput.Focus()
		m.screen = ScreenNameEditing
		return m, nil
	}

	var cmd tea.Cmd
	m.remoteInput, cmd = m.remoteInput.Update(msg)
	return m, cmd
}

func (m Model) updateNameEditing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		m.cancel()
		return m, tea.Quit

	case "d":
		m.debugMode = !m.debugMode

	case "esc":
		// If using custom remote, go back to custom remote screen
		if m.customRemote != "" {
			m.remoteInput.Focus()
			m.screen = ScreenCustomRemote
		} else {
			m.screen = ScreenPresetSelection
		}
		return m, nil

	case "enter":
		m.archiveName = m.textInput.Value()
		if m.archiveName == "" {
			return m, nil // Don't proceed with empty name
		}
		
		// Determine remote path
		var remotePath string
		if m.customRemote != "" {
			remotePath = m.customRemote
		} else {
			preset := m.presets[m.selectedIndex]
			remotePath = preset.Remote
		}
		
		// Start archiving
		uploadID, err := archive.Archive(m.ctx, m.folderPath, remotePath, m.archiveName, m.debugMode)
		if err != nil {
			m.err = err
			return m, tea.Quit
		}
		
		// Store upload ID and quit to launch observe
		m.uploadID = uploadID
		m.quitting = true
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View renders the UI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.screen {
	case ScreenPresetSelection:
		return m.viewPresetSelection()
	case ScreenCustomRemote:
		return m.viewCustomRemote()
	case ScreenNameEditing:
		return m.viewNameEditing()
	}

	return ""
}

func (m Model) viewPresetSelection() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	b.WriteString(titleStyle.Render("📦 Jotta Archiver"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Select a preset for: %s\n\n", m.folderPath))

	// Display all presets with consistent formatting
	for i, preset := range m.presets {
		cursor := " "
		if i == m.selectedIndex {
			cursor = ">"
		}
		
		// Format: "> Preset Name" or "  Preset Name"
		line := fmt.Sprintf("%s %s", cursor, preset.Name)
		
		if i == m.selectedIndex {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
		
		// Show remote path indented below
		remoteLine := fmt.Sprintf("    %s", preset.Remote)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(remoteLine))
		b.WriteString("\n")
	}
	
	// Add "Custom..." option
	cursor := " "
	if m.selectedIndex == len(m.presets) {
		cursor = ">"
	}
	
	customLine := fmt.Sprintf("%s Custom...", cursor)
	if m.selectedIndex == len(m.presets) {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render(customLine))
	} else {
		b.WriteString(customLine)
	}
	b.WriteString("\n")
	
	customDescLine := "    Enter custom remote path"
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(customDescLine))
	b.WriteString("\n")

	b.WriteString("\n")
	helpText := "↑/↓: Navigate • Enter: Select • d: Toggle Debug • q: Quit"
	if m.debugMode {
		helpText = "↑/↓: Navigate • Enter: Select • d: Toggle Debug (ON) • q: Quit"
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(helpText))

	return b.String()
}

func (m Model) viewCustomRemote() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	b.WriteString(titleStyle.Render("🔧 Custom Remote Path"))
	b.WriteString("\n\n")

	b.WriteString("Enter remote directory path:\n")
	b.WriteString(m.remoteInput.View())
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Enter: Continue • Esc: Back • Ctrl+C: Quit"))

	return b.String()
}

func (m Model) viewNameEditing() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	b.WriteString(titleStyle.Render("📝 Archive Name"))
	b.WriteString("\n\n")

	if m.customRemote != "" {
		b.WriteString("Remote: " + m.customRemote + "\n\n")
	} else {
		preset := m.presets[m.selectedIndex]
		b.WriteString(fmt.Sprintf("Preset: %s\n", preset.Name))
		b.WriteString(fmt.Sprintf("Remote: %s\n\n", preset.Remote))
	}

	b.WriteString("Archive name:\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	helpText := "Enter: Start • Esc: Back • d: Toggle Debug • Ctrl+C: Quit"
	if m.debugMode {
		helpText = "Enter: Start • Esc: Back • d: Toggle Debug (ON) • Ctrl+C: Quit"
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(helpText))

	return b.String()
}

// Run starts the TUI application and returns the upload ID (if upload started) and debug mode
func Run(presets []config.Preset, initialName, folderPath string) (string, bool, error) {
	m := New(presets, initialName, folderPath)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", false, err
	}
	
	// Extract upload ID and debug mode from final model
	if fm, ok := finalModel.(Model); ok {
		return fm.uploadID, fm.debugMode, nil
	}
	
	return "", false, nil
}

