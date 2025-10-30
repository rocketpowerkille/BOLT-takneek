package UI

import (
	"fmt"
	"os"
	"strings"

	"github.com/pclubiitk/dbcli/DB"
)

func (m Model) View() string {
	var baseView string
	
	switch m.Step {
	case StepSourceCred, StepDestCred:
		baseView = ViewDBCred(m)
	case StepSelectSourceTable, StepSelectSourceColumns, StepSelectDestTable, StepSelectDestColumns:
		baseView = ViewSelection(m)
	case StepMapping:
		baseView = ViewMapping(m)
	case StepDumpOption:
		baseView = ViewDumpOption(m)
	case StepMigrationConfirm:
		baseView = ViewMigrationConfirm(m)
	default:
		baseView = ViewMigrationComplete(m)
	}
	
	// Add help overlay if active
	return AddHelpToView(baseView, m)
}

func ViewMigrationConfirm(m Model) string {
	var s strings.Builder
	
	s.WriteString("Migration Summary\n")
	s.WriteString("=================\n\n")
	
	s.WriteString(fmt.Sprintf("Source: %s.%s\n", m.SourceCred["dbVendor"], m.SelectedSourceTbl))
	s.WriteString(fmt.Sprintf("Destination: %s.%s\n", m.DestCred["dbVendor"], m.SelectedDestTbl))
	s.WriteString("\nColumn Mappings:\n")
	for src, dest := range m.ColumnMapping {
		s.WriteString(fmt.Sprintf("  %s → %s\n", src, dest))
	}
	
	if m.WantDump {
		s.WriteString(fmt.Sprintf("\nSQL dump will be saved to: %s\n", m.DumpPath))
	}
	
	// Show processing status if migration is in progress
	if m.IsProcessing {
		s.WriteString("\n" + strings.Repeat("=", 50) + "\n")
		s.WriteString(fmt.Sprintf("🔄 %s\n", m.ProcessingMsg))
		
		if m.TotalRows > 0 {
			// Draw progress bar
			percentage := float64(m.ProcessedRows) / float64(m.TotalRows)
			barWidth := 40
			filledWidth := int(percentage * float64(barWidth))
			
			s.WriteString("\nProgress: [")
			s.WriteString(strings.Repeat("█", filledWidth))
			s.WriteString(strings.Repeat("░", barWidth-filledWidth))
			s.WriteString(fmt.Sprintf("] %.1f%%\n", percentage*100))
			s.WriteString(fmt.Sprintf("Rows: %d/%d\n", m.ProcessedRows, m.TotalRows))
		}
		s.WriteString("\nPlease wait...")
	} else {
		s.WriteString("\nPress 'y' to start migration, 'n' to cancel, 'd' to generate dump only\n")
	}
	
	if m.ErrMsg != "" {
		s.WriteString(fmt.Sprintf("\n❌ Error: %s\n", m.ErrMsg))
	}
	
	return s.String()
}

func ViewMigrationComplete(m Model) string {
	var s strings.Builder
	
	s.WriteString("Migration Complete!\n")
	s.WriteString("===================\n\n")
	
	s.WriteString("Successfully migrated data with the following mappings:\n")
	for src, dest := range m.ColumnMapping {
		s.WriteString(fmt.Sprintf("  %s → %s\n", src, dest))
	}
	
	if m.WantDump && m.DumpPath != "" {
		s.WriteString(fmt.Sprintf("\nSQL dump saved to: %s\n", m.DumpPath))
	}
	
	s.WriteString("\nPress 'q' to quit.\n")
	
	return s.String()
}

func executeMigration(m Model) Model {
	if m.Source == nil || m.Dest == nil {
		m.ErrMsg = "Source or destination database not connected"
		return m
	}

	// Set processing state
	m.IsProcessing = true
	m.ProcessingMsg = "Starting migration..."
	m.ProcessedRows = 0
	m.TotalRows = 0

	// Create progress callback
	progressCallback := func(processed, total int) {
		m.ProcessedRows = processed
		m.TotalRows = total
		percentage := float64(processed) / float64(total) * 100
		m.ProcessingMsg = fmt.Sprintf("Migrating data... %.1f%% (%d/%d rows)", percentage, processed, total)
	}

	// Perform the migration with progress tracking
	err := DB.MigrateDataWithProgress(m.Source, m.Dest, m.SelectedSourceTbl, m.SelectedDestTbl, m.ColumnMapping, progressCallback)
	if err != nil {
		m.ErrMsg = fmt.Sprintf("Migration failed: %v", err)
		m.IsProcessing = false
		return m
	}

	// Generate dump if requested
	if m.WantDump {
		m.ProcessingMsg = "Generating SQL dump..."
		dump, err := DB.GenerateSQLDump(m.Source, m.SelectedSourceTbl, m.ColumnMapping, m.SelectedDestTbl)
		if err != nil {
			m.ErrMsg = fmt.Sprintf("Failed to generate dump: %v", err)
			m.IsProcessing = false
			return m
		}

		err = os.WriteFile(m.DumpPath, []byte(dump), 0644)
		if err != nil {
			m.ErrMsg = fmt.Sprintf("Failed to save dump file: %v", err)
			m.IsProcessing = false
			return m
		}
	}

	m.Step++
	m.ErrMsg = ""
	m.IsProcessing = false
	m.ProcessingMsg = "Migration completed successfully!"
	return m
}

func generateDumpOnly(m Model) Model {
	if m.Source == nil {
		m.ErrMsg = "Source database not connected"
		return m
	}

	dump, err := DB.GenerateSQLDump(m.Source, m.SelectedSourceTbl, m.ColumnMapping, m.SelectedDestTbl)
	if err != nil {
		m.ErrMsg = fmt.Sprintf("Failed to generate dump: %v", err)
		return m
	}

	err = os.WriteFile(m.DumpPath, []byte(dump), 0644)
	if err != nil {
		m.ErrMsg = fmt.Sprintf("Failed to save dump file: %v", err)
		return m
	}

	m.Step++
	m.ErrMsg = ""
	return m
}
