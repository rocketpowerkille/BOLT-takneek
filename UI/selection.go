package UI

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pclubiitk/dbcli/DB"
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func UpdateSelection(m Model, msg tea.Msg) Model {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch m.Step {
			case StepSelectSourceTable:
				// Fetch tables from source database if not already done
				if len(m.SourceTables) == 0 {
					m = fetchSourceTables(m)
				}
				if len(m.SourceTables) > 0 {
					if selectedItem, ok := m.SourceTableList.SelectedItem().(item); ok {
						m.SelectedSourceTbl = selectedItem.title
					} else if len(m.SourceTables) > 0 {
						m.SelectedSourceTbl = m.SourceTables[0]
					}
					m.Step++
				}
			case StepSelectSourceColumns:
				// Fetch columns for selected source table
				if len(m.SourceColumns) == 0 && m.SelectedSourceTbl != "" {
					m = fetchSourceColumns(m)
				}
				// Get selected columns from selections map
				if len(m.SourceColumns) > 0 {
					m.SelectedSourceCols = getSelectedColumnsFromMap(m.SourceColumns, m.SourceColSelections)
					if len(m.SelectedSourceCols) == 0 {
						m.ErrMsg = "Please select at least one column"
						return m
					}
					m.Step++
				}
			case StepSelectDestTable:
				// Fetch tables from destination database if not already done
				if len(m.DestTables) == 0 {
					m = fetchDestTables(m)
				}
				if len(m.DestTables) > 0 {
					if selectedItem, ok := m.DestTableList.SelectedItem().(item); ok {
						m.SelectedDestTbl = selectedItem.title
					} else if len(m.DestTables) > 0 {
						m.SelectedDestTbl = m.DestTables[0]
					}
					m.Step++
				}
			case StepSelectDestColumns:
				// Fetch columns for selected destination table
				if len(m.DestColumns) == 0 && m.SelectedDestTbl != "" {
					m = fetchDestColumns(m)
				}
				// Get selected columns from selections map
				if len(m.DestColumns) > 0 {
					m.SelectedDestCols = getSelectedColumnsFromMap(m.DestColumns, m.DestColSelections)
					if len(m.SelectedDestCols) == 0 {
						m.ErrMsg = "Please select at least one column"
						return m
					}
					m.Step++
				}
			}
		case " ": // Space to toggle selection
			switch m.Step {
			case StepSelectSourceColumns:
				if m.SourceColSelections == nil {
					m.SourceColSelections = make(map[int]bool)
				}
				selectedIdx := m.SourceColumnList.Index()
				m.SourceColSelections[selectedIdx] = !m.SourceColSelections[selectedIdx]
				m.ErrMsg = "" // Clear error when user makes selection
			case StepSelectDestColumns:
				if m.DestColSelections == nil {
					m.DestColSelections = make(map[int]bool)
				}
				selectedIdx := m.DestColumnList.Index()
				m.DestColSelections[selectedIdx] = !m.DestColSelections[selectedIdx]
				m.ErrMsg = "" // Clear error when user makes selection
			}
		case "a": // Select all columns
			switch m.Step {
			case StepSelectSourceColumns:
				if m.SourceColSelections == nil {
					m.SourceColSelections = make(map[int]bool)
				}
				for i := range m.SourceColumns {
					m.SourceColSelections[i] = true
				}
				m.ErrMsg = ""
			case StepSelectDestColumns:
				if m.DestColSelections == nil {
					m.DestColSelections = make(map[int]bool)
				}
				for i := range m.DestColumns {
					m.DestColSelections[i] = true
				}
				m.ErrMsg = ""
			}
		case "n": // Select none
			switch m.Step {
			case StepSelectSourceColumns:
				m.SourceColSelections = make(map[int]bool)
				m.ErrMsg = ""
			case StepSelectDestColumns:
				m.DestColSelections = make(map[int]bool)
				m.ErrMsg = ""
			}
		}
	}

	// Update the lists for navigation
	switch m.Step {
	case StepSelectSourceTable:
		var cmd tea.Cmd
		m.SourceTableList, cmd = m.SourceTableList.Update(msg)
		_ = cmd
	case StepSelectSourceColumns:
		var cmd tea.Cmd
		m.SourceColumnList, cmd = m.SourceColumnList.Update(msg)
		_ = cmd
	case StepSelectDestTable:
		var cmd tea.Cmd
		m.DestTableList, cmd = m.DestTableList.Update(msg)
		_ = cmd
	case StepSelectDestColumns:
		var cmd tea.Cmd
		m.DestColumnList, cmd = m.DestColumnList.Update(msg)
		_ = cmd
	}

	return m
}

func fetchSourceTables(m Model) Model {
	if m.Source == nil {
		m.ErrMsg = "Source database not connected"
		return m
	}

	dbType := getDBType(m.SourceCred["dbVendor"])
	tables, err := DB.FetchTablesGeneric(m.Source, dbType)
	if err != nil {
		m.ErrMsg = fmt.Sprintf("Error fetching source tables: %v", err)
		return m
	}

	m.SourceTables = tables
	m.SourceTableList = createTableList(tables)
	m.ErrMsg = ""
	return m
}

func fetchDestTables(m Model) Model {
	if m.Dest == nil {
		m.ErrMsg = "Destination database not connected"
		return m
	}

	dbType := getDBType(m.DestCred["dbVendor"])
	tables, err := DB.FetchTablesGeneric(m.Dest, dbType)
	if err != nil {
		m.ErrMsg = fmt.Sprintf("Error fetching destination tables: %v", err)
		return m
	}

	m.DestTables = tables
	m.DestTableList = createTableList(tables)
	m.ErrMsg = ""
	return m
}

func fetchSourceColumns(m Model) Model {
	if m.Source == nil || m.SelectedSourceTbl == "" {
		m.ErrMsg = "Source database not connected or no table selected"
		return m
	}

	dbType := getDBType(m.SourceCred["dbVendor"])
	columns, err := DB.FetchColumnsGeneric(m.Source, m.SelectedSourceTbl, dbType)
	if err != nil {
		m.ErrMsg = fmt.Sprintf("Error fetching source columns: %v", err)
		return m
	}

	m.SourceColumns = columns
	m.SourceColumnList = createColumnList(columns)
	m.ErrMsg = ""
	return m
}

func fetchDestColumns(m Model) Model {
	if m.Dest == nil || m.SelectedDestTbl == "" {
		m.ErrMsg = "Destination database not connected or no table selected"
		return m
	}

	dbType := getDBType(m.DestCred["dbVendor"])
	columns, err := DB.FetchColumnsGeneric(m.Dest, m.SelectedDestTbl, dbType)
	if err != nil {
		m.ErrMsg = fmt.Sprintf("Error fetching destination columns: %v", err)
		return m
	}

	m.DestColumns = columns
	m.DestColumnList = createColumnList(columns)
	m.ErrMsg = ""
	return m
}

func getDBType(vendor string) string {
	return strings.ToLower(vendor)
}

func createTableList(tables []string) list.Model {
	items := make([]list.Item, len(tables))
	for i, table := range tables {
		items[i] = item{title: table, desc: fmt.Sprintf("Table: %s", table)}
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a table"
	return l
}

func createColumnList(columns []string) list.Model {
	items := make([]list.Item, len(columns))
	for i, column := range columns {
		items[i] = item{title: column, desc: fmt.Sprintf("Column: %s", column)}
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select columns (Space to toggle, Enter to continue)"
	return l
}

func getSelectedColumns(columnList list.Model, allColumns []string) []string {
	// For now, return all columns. In a full implementation,
	// you'd track which items are selected
	return allColumns
}

func getSelectedColumnsFromMap(allColumns []string, selections map[int]bool) []string {
	var selected []string
	for i, column := range allColumns {
		if selections[i] {
			selected = append(selected, column)
		}
	}
	return selected
}

func ViewSelection(m Model) string {
	var s strings.Builder
	
	switch m.Step {
	case StepSelectSourceTable:
		if len(m.SourceTables) == 0 {
			s.WriteString("Fetching source tables...\n")
		} else {
			s.WriteString("Select source table:\n")
			s.WriteString("====================\n\n")
			s.WriteString(m.SourceTableList.View())
			s.WriteString("\nUse ↑↓ to navigate, Enter to select")
		}
	case StepSelectSourceColumns:
		if len(m.SourceColumns) == 0 {
			s.WriteString(fmt.Sprintf("Fetching columns for table '%s'...\n", m.SelectedSourceTbl))
		} else {
			s.WriteString(fmt.Sprintf("Select columns from table '%s':\n", m.SelectedSourceTbl))
			s.WriteString("====================================\n\n")
			
			// Show columns with selection status
			for i, column := range m.SourceColumns {
				marker := "  "
				if m.SourceColSelections != nil && m.SourceColSelections[i] {
					marker = "✓ "
				}
				
				// Highlight current selection
				if i == m.SourceColumnList.Index() {
					s.WriteString(fmt.Sprintf("► %s%s\n", marker, column))
				} else {
					s.WriteString(fmt.Sprintf("  %s%s\n", marker, column))
				}
			}
			
			selectedCount := len(getSelectedColumnsFromMap(m.SourceColumns, m.SourceColSelections))
			s.WriteString(fmt.Sprintf("\nSelected: %d/%d columns\n", selectedCount, len(m.SourceColumns)))
			s.WriteString("\nControls: ↑↓ navigate, Space to toggle, 'a' select all, 'n' select none, Enter to continue")
		}
	case StepSelectDestTable:
		if len(m.DestTables) == 0 {
			s.WriteString("Fetching destination tables...\n")
		} else {
			s.WriteString("Select destination table:\n")
			s.WriteString("=========================\n\n")
			s.WriteString(m.DestTableList.View())
			s.WriteString("\nUse ↑↓ to navigate, Enter to select")
		}
	case StepSelectDestColumns:
		if len(m.DestColumns) == 0 {
			s.WriteString(fmt.Sprintf("Fetching columns for table '%s'...\n", m.SelectedDestTbl))
		} else {
			s.WriteString(fmt.Sprintf("Select columns from table '%s':\n", m.SelectedDestTbl))
			s.WriteString("====================================\n\n")
			
			// Show columns with selection status
			for i, column := range m.DestColumns {
				marker := "  "
				if m.DestColSelections != nil && m.DestColSelections[i] {
					marker = "✓ "
				}
				
				// Highlight current selection
				if i == m.DestColumnList.Index() {
					s.WriteString(fmt.Sprintf("► %s%s\n", marker, column))
				} else {
					s.WriteString(fmt.Sprintf("  %s%s\n", marker, column))
				}
			}
			
			selectedCount := len(getSelectedColumnsFromMap(m.DestColumns, m.DestColSelections))
			s.WriteString(fmt.Sprintf("\nSelected: %d/%d columns\n", selectedCount, len(m.DestColumns)))
			s.WriteString("\nControls: ↑↓ navigate, Space to toggle, 'a' select all, 'n' select none, Enter to continue")
		}
	}

	if m.ErrMsg != "" {
		s.WriteString(fmt.Sprintf("\n\n❌ Error: %s", m.ErrMsg))
	}

	s.WriteString("\n\nNavigation:")
	s.WriteString("\n  Ctrl+B - Go back to previous step")
	s.WriteString("\n  Ctrl+C - Quit application")

	return s.String()
}
