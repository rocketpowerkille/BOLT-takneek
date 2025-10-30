package UI

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func UpdateMapping(m Model, msg tea.Msg) Model {
	if m.ColumnMapping == nil {
		m.ColumnMapping = make(map[string]string)
		m.CurrentMapIdx = 0
		m.MapInput = textinput.New()
		m.MapInput.Placeholder = "Enter destination column name or press Enter to auto-map"
		m.MapInput.Focus()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.CurrentMapIdx < len(m.SelectedSourceCols) {
				sourceCol := m.SelectedSourceCols[m.CurrentMapIdx]
				destCol := m.MapInput.Value()
				
				// Auto-map if no input provided
				if destCol == "" {
					if m.CurrentMapIdx < len(m.SelectedDestCols) {
						destCol = m.SelectedDestCols[m.CurrentMapIdx]
					} else {
						destCol = sourceCol // Use same name if no dest column available
					}
				}
				
				// Validate that destination column exists
				if !contains(m.SelectedDestCols, destCol) {
					m.ErrMsg = fmt.Sprintf("Column '%s' not found in destination table", destCol)
					return m
				}
				
				m.ColumnMapping[sourceCol] = destCol
				m.CurrentMapIdx++
				m.MapInput.SetValue("")
				m.ErrMsg = ""
				
				// If we've mapped all columns, proceed to next step
				if m.CurrentMapIdx >= len(m.SelectedSourceCols) {
					m.Step++
				}
			}
		case "ctrl+c", "esc":
			// Skip current mapping
			if m.CurrentMapIdx < len(m.SelectedSourceCols) {
				m.CurrentMapIdx++
				m.MapInput.SetValue("")
				if m.CurrentMapIdx >= len(m.SelectedSourceCols) {
					m.Step++
				}
			}
		default:
			var cmd tea.Cmd
			m.MapInput, cmd = m.MapInput.Update(msg)
			_ = cmd
		}
	}
	
	return m
}

func ViewMapping(m Model) string {
	var s strings.Builder
	
	s.WriteString("Column Mapping Configuration\n")
	s.WriteString("============================\n\n")
	
	// Show existing mappings
	if len(m.ColumnMapping) > 0 {
		s.WriteString("Current Mappings:\n")
		for src, dest := range m.ColumnMapping {
			s.WriteString(fmt.Sprintf("  %s → %s\n", src, dest))
		}
		s.WriteString("\n")
	}
	
	// Show current mapping in progress
	if m.CurrentMapIdx < len(m.SelectedSourceCols) {
		sourceCol := m.SelectedSourceCols[m.CurrentMapIdx]
		s.WriteString(fmt.Sprintf("Mapping source column: %s\n", sourceCol))
		s.WriteString("Available destination columns:\n")
		for i, col := range m.SelectedDestCols {
			marker := "  "
			if i == m.CurrentMapIdx && m.CurrentMapIdx < len(m.SelectedDestCols) {
				marker = "→ "
			}
			s.WriteString(fmt.Sprintf("%s%s\n", marker, col))
		}
		s.WriteString("\n")
		s.WriteString(fmt.Sprintf("Destination column: %s\n", m.MapInput.View()))
		s.WriteString("Press Enter to confirm mapping, Esc to skip\n")
		s.WriteString("Leave empty to auto-map to suggested column\n")
	} else {
		s.WriteString("All columns mapped! Press Enter to continue.\n")
	}
	
	if m.ErrMsg != "" {
		s.WriteString(fmt.Sprintf("\nError: %s\n", m.ErrMsg))
	}
	
	return s.String()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
