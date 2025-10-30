package UI

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func UpdateDumpOption(m Model, msg tea.Msg) Model {
	// Initialize dump path input if not done
	if m.DumpPathInp.Value() == "" {
		m.DumpPathInp = textinput.New()
		m.DumpPathInp.Placeholder = "./migration_dump.sql"
		m.DumpPathInp.SetValue("./migration_dump.sql")
		m.DumpPathInp.Focus()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.DumpPath = m.DumpPathInp.Value()
			if m.DumpPath == "" {
				m.DumpPath = "./migration_dump.sql"
			}
			m.WantDump = true
			m.Step++
		case "ctrl+s", "s":
			// Skip dump option
			m.WantDump = false
			m.DumpPath = ""
			m.Step++
		default:
			var cmd tea.Cmd
			m.DumpPathInp, cmd = m.DumpPathInp.Update(msg)
			_ = cmd
		}
	}
	return m
}

func ViewDumpOption(m Model) string {
	var s strings.Builder
	
	s.WriteString("SQL Dump Configuration\n")
	s.WriteString("======================\n\n")
	
	s.WriteString("Would you like to generate an SQL dump file?\n")
	s.WriteString("This will create a .sql file with INSERT statements for the migration.\n\n")
	
	s.WriteString(fmt.Sprintf("Dump file path: %s\n", m.DumpPathInp.View()))
	s.WriteString("\nPress Enter to enable dump generation")
	s.WriteString("\nPress 's' to skip dump generation")
	
	if m.ErrMsg != "" {
		s.WriteString(fmt.Sprintf("\nError: %s", m.ErrMsg))
	}
	
	return s.String()
}
