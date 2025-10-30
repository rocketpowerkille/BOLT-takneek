package UI

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle help first
	m = UpdateHelp(m, msg)
	
	// If help is visible and we handled a key, don't process other keys
	if m.HelpState != nil && m.HelpState.Visible {
		// Only allow help and exit keys when help is visible
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "h", "?", "esc":
				// Already handled in UpdateHelp
				return m, nil
			default:
				// Ignore other keys when help is visible
				return m, nil
			}
		}
		return m, nil
	}

	// Global key handling
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+b", "b": // Go back to previous step
			if !m.IsProcessing { // Don't allow going back during processing
				m = goBackStep(m)
				return m, nil
			}
		}
	}

	// Step-specific handling
	switch m.Step {
	case StepSourceCred, StepDestCred:
		m = UpdateDBCred(m, msg)
	case StepSelectSourceTable, StepSelectSourceColumns, StepSelectDestTable, StepSelectDestColumns:
		m = UpdateSelection(m, msg)
	case StepMapping:
		m = UpdateMapping(m, msg)
	case StepDumpOption:
		m = UpdateDumpOption(m, msg)
	case StepMigrationConfirm:
		if !m.IsProcessing { // Only handle input if not processing
			m = UpdateMigrationConfirm(m, msg)
		}
	}

	return m, nil
}

func goBackStep(m Model) Model {
	// Clear any error messages
	m.ErrMsg = ""
	
	switch m.Step {
	case StepDestCred:
		// Go back to source credentials
		m.Step = StepSourceCred
		m.IsSource = true
		m.CredIndex = 0
		m.CredInput.SetValue("")
	case StepSelectSourceTable:
		// Go back to destination credentials
		m.Step = StepDestCred
		m.IsSource = false
		m.CredIndex = 0
		m.CredInput.SetValue("")
	case StepSelectSourceColumns:
		// Go back to source table selection
		m.Step = StepSelectSourceTable
		m.SelectedSourceTbl = ""
		m.SourceColumns = nil
		m.SelectedSourceCols = nil
		m.SourceColSelections = nil
	case StepSelectDestTable:
		// Go back to source column selection
		m.Step = StepSelectSourceColumns
		m.DestTables = nil
		m.SelectedDestTbl = ""
	case StepSelectDestColumns:
		// Go back to destination table selection
		m.Step = StepSelectDestTable
		m.SelectedDestTbl = ""
		m.DestColumns = nil
		m.SelectedDestCols = nil
		m.DestColSelections = nil
	case StepMapping:
		// Go back to destination column selection
		m.Step = StepSelectDestColumns
		m.ColumnMapping = nil
		m.CurrentMapIdx = 0
	case StepDumpOption:
		// Go back to column mapping
		m.Step = StepMapping
		m.WantDump = false
		m.DumpPath = ""
	case StepMigrationConfirm:
		// Go back to dump option
		m.Step = StepDumpOption
	}
	
	return m
}

func UpdateMigrationConfirm(m Model, msg tea.Msg) Model {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			// Execute full migration
			m = executeMigration(m)
		case "d", "D":
			// Generate dump only
			m = generateDumpOnly(m)
		case "n", "N":
			// Cancel - go back to previous step
			m.Step = StepDumpOption
		}
	}
	return m
}
