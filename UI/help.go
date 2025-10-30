package UI

import (
	"strings"
	tea "github.com/charmbracelet/bubbletea"
)

// UpdateHelp handles help-related key presses
func UpdateHelp(m Model, msg tea.Msg) Model {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "h", "?":
			// Toggle help visibility
			if m.HelpState == nil {
				m.HelpState = &HelpState{}
			}
			m.HelpState.Visible = !m.HelpState.Visible
			m.HelpState.CurrentStep = m.Step
		case "esc":
			// Close help if it's open
			if m.HelpState != nil && m.HelpState.Visible {
				m.HelpState.Visible = false
				return m
			}
		}
	}
	return m
}

// ViewHelp returns the help content for the current step
func ViewHelp(m Model) string {
	if m.HelpState == nil || !m.HelpState.Visible {
		return ""
	}

	var s strings.Builder
	
	s.WriteString("╔════════════════════════════════════════════════════════════════════════╗\n")
	s.WriteString("║                                HELP SYSTEM                            ║\n")
	s.WriteString("╠════════════════════════════════════════════════════════════════════════╣\n")
	
	switch m.Step {
	case StepSourceCred, StepDestCred:
		s.WriteString(getCredentialsHelp())
	case StepSelectSourceTable, StepSelectDestTable:
		s.WriteString(getTableSelectionHelp())
	case StepSelectSourceColumns, StepSelectDestColumns:
		s.WriteString(getColumnSelectionHelp())
	case StepMapping:
		s.WriteString(getMappingHelp())
	case StepDumpOption:
		s.WriteString(getDumpOptionHelp())
	case StepMigrationConfirm:
		s.WriteString(getMigrationConfirmHelp())
	default:
		s.WriteString(getGeneralHelp())
	}
	
	s.WriteString("╠════════════════════════════════════════════════════════════════════════╣\n")
	s.WriteString("║ Press 'h' or '?' to toggle help | Press 'Esc' to close help           ║\n")
	s.WriteString("╚════════════════════════════════════════════════════════════════════════╝\n")
	
	return s.String()
}

func getCredentialsHelp() string {
	return `║ DATABASE CREDENTIALS HELP                                             ║
║                                                                        ║
║ This step collects connection information for your database.           ║
║                                                                        ║
║ Fields explained:                                                      ║
║ • Database Vendor: Type 'mysql' or 'oracle'                          ║
║ • Host: Server hostname (localhost, IP address, or domain name)       ║
║ • Port: Database port number                                          ║
║   - MySQL default: 3306                                               ║
║   - Oracle default: 1521                                              ║
║ • Username: Database user account                                     ║
║ • Password: Database password (input will be masked)                  ║
║ • Database: Database name (MySQL) or Service name (Oracle)            ║
║                                                                        ║
║ Navigation:                                                            ║
║ • Type your input and press Enter to move to next field               ║
║ • Press Esc to go back to previous field                              ║
║ • Connection will be tested automatically after all fields entered    ║
║                                                                        ║
║ Tips:                                                                  ║
║ • For Oracle, use service name (like XE, ORCL) not SID               ║
║ • Ensure your database server is running and accessible              ║
║ • Check firewall settings if connection fails                         ║`
}

func getTableSelectionHelp() string {
	return `║ TABLE SELECTION HELP                                                  ║
║                                                                        ║
║ Select the table you want to migrate data from/to.                    ║
║                                                                        ║
║ Features:                                                              ║
║ • Tables are loaded automatically from your connected database         ║
║ • Use arrow keys (↑↓) to navigate through the list                   ║
║ • Press Enter to select the highlighted table                         ║
║                                                                        ║
║ Navigation:                                                            ║
║ • ↑↓ Arrow Keys: Move up/down in the list                            ║
║ • Enter: Select the current table and proceed                         ║
║ • Ctrl+B: Go back to previous step                                    ║
║ • Ctrl+C: Quit application                                            ║
║                                                                        ║
║ Troubleshooting:                                                       ║
║ • If no tables appear, check your database connection                 ║
║ • Ensure your user has SELECT permissions on the tables               ║
║ • For Oracle, only user tables are shown (not system tables)         ║
║                                                                        ║
║ Note: The table list shows all tables you have access to read         ║`
}

func getColumnSelectionHelp() string {
	return `║ COLUMN SELECTION HELP                                                 ║
║                                                                        ║
║ Choose which columns to include in the migration.                     ║
║                                                                        ║
║ Features:                                                              ║
║ • Individual column selection with checkboxes                         ║
║ • Bulk selection options for convenience                              ║
║ • Visual indicators show selected columns                             ║
║                                                                        ║
║ Controls:                                                              ║
║ • ↑↓ Arrow Keys: Navigate through columns                             ║
║ • Space: Toggle selection of current column                           ║
║ • 'a' Key: Select ALL columns at once                                ║
║ • 'n' Key: Select NONE (deselect all columns)                        ║
║ • Enter: Confirm selection and proceed                                ║
║                                                                        ║
║ Visual Indicators:                                                     ║
║ • ✓ : Column is selected for migration                               ║
║ • ► : Currently highlighted column                                    ║
║ • Selected count is shown at the bottom                              ║
║                                                                        ║
║ Tips:                                                                  ║
║ • Select at least one column to proceed                              ║
║ • Consider data types when selecting columns                         ║
║ • Identity/auto-increment columns may need special handling          ║`
}

func getMappingHelp() string {
	return `║ COLUMN MAPPING HELP                                                   ║
║                                                                        ║
║ Map source columns to destination columns.                            ║
║                                                                        ║
║ Process:                                                               ║
║ • Each source column needs to be mapped to a destination column       ║
║ • Auto-mapping suggests the corresponding column                      ║
║ • You can override suggestions with custom mappings                   ║
║                                                                        ║
║ Controls:                                                              ║
║ • Type destination column name (or leave blank for auto-mapping)      ║
║ • Enter: Confirm mapping and move to next column                      ║
║ • Esc: Skip current mapping                                           ║
║                                                                        ║
║ Auto-mapping:                                                          ║
║ • If you press Enter without typing, auto-mapping is used            ║
║ • Auto-mapping matches column names or uses positional mapping       ║
║ • Review suggestions carefully before accepting                       ║
║                                                                        ║
║ Validation:                                                            ║
║ • Destination column must exist in the target table                  ║
║ • Data type compatibility will be checked                            ║
║ • Warnings will be shown for potential issues                        ║
║                                                                        ║
║ Best Practices:                                                        ║
║ • Map similar data types together                                     ║
║ • Consider NULL constraints and data sizes                            ║`
}

func getDumpOptionHelp() string {
	return `║ SQL DUMP OPTION HELP                                                  ║
║                                                                        ║
║ Generate SQL INSERT statements for the migration.                     ║
║                                                                        ║
║ What is a SQL Dump?                                                    ║
║ • A text file containing SQL INSERT statements                        ║
║ • Can be used to recreate the migration manually                      ║
║ • Useful for review, backup, or manual execution                      ║
║                                                                        ║
║ Options:                                                               ║
║ • Enter: Generate dump file at specified path                         ║
║ • 's' Key: Skip dump generation and proceed to migration             ║
║                                                                        ║
║ File Path:                                                             ║
║ • Default: ./migration_dump.sql                                       ║
║ • You can customize the path and filename                             ║
║ • Ensure you have write permissions to the directory                  ║
║                                                                        ║
║ Use Cases:                                                             ║
║ • Review migration before execution                                    ║
║ • Keep backup of migration commands                                    ║
║ • Execute migration manually in parts                                 ║
║ • Share migration with team for review                                ║
║                                                                        ║
║ Note: Dump generation is optional and doesn't affect migration        ║`
}

func getMigrationConfirmHelp() string {
	return `║ MIGRATION CONFIRMATION HELP                                           ║
║                                                                        ║
║ Final review and execution of your database migration.                ║
║                                                                        ║
║ Summary Information:                                                   ║
║ • Source and destination databases and tables                         ║
║ • Column mappings that will be applied                                ║
║ • SQL dump file location (if enabled)                                 ║
║                                                                        ║
║ Options:                                                               ║
║ • 'y' or 'Y': Execute full migration with data transfer              ║
║ • 'd' or 'D': Generate dump file only (no data transfer)             ║
║ • 'n' or 'N': Cancel and go back to previous step                    ║
║                                                                        ║
║ Migration Process:                                                     ║
║ • Data is transferred in batches for better performance               ║
║ • Progress bar shows completion status                                ║
║ • Transaction support ensures data integrity                          ║
║ • Process can take time depending on data volume                      ║
║                                                                        ║
║ Safety Features:                                                       ║
║ • Transaction rollback on errors                                      ║
║ • Batch processing prevents memory issues                             ║
║ • Detailed logging of all operations                                  ║
║                                                                        ║
║ Warning: This will modify your destination database!                  ║`
}

func getGeneralHelp() string {
	return `║ GENERAL HELP                                                          ║
║                                                                        ║
║ Welcome to DBCLI - Database Migration Tool                           ║
║                                                                        ║
║ Global Controls:                                                       ║
║ • 'h' or '?': Show/hide help for current step                        ║
║ • Ctrl+B or 'b': Go back to previous step                            ║
║ • Ctrl+C or 'q': Quit application                                    ║
║ • Esc: Close help or cancel current operation                        ║
║                                                                        ║
║ Migration Steps:                                                       ║
║ 1. Source Database Credentials                                        ║
║ 2. Destination Database Credentials                                   ║
║ 3. Source Table Selection                                             ║
║ 4. Source Column Selection                                            ║
║ 5. Destination Table Selection                                        ║
║ 6. Destination Column Selection                                       ║
║ 7. Column Mapping Configuration                                       ║
║ 8. SQL Dump Option                                                    ║
║ 9. Migration Confirmation & Execution                                 ║
║                                                                        ║
║ Features:                                                              ║
║ • Support for MySQL and Oracle databases                             ║
║ • Transaction-safe migrations                                         ║
║ • Progress tracking and batch processing                              ║
║ • Comprehensive data validation                                       ║
║ • Connection pooling for better performance                           ║`
}

// AddHelpToView adds help overlay to any view if help is active
func AddHelpToView(baseView string, m Model) string {
	if m.HelpState == nil || !m.HelpState.Visible {
		// Add help hint at the bottom
		return baseView + "\n\nPress 'h' or '?' for help"
	}
	
	// Show help overlay
	helpContent := ViewHelp(m)
	
	// Combine base view with help overlay
	return baseView + "\n\n" + helpContent
}
