package preview

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/pclubiitk/dbcli/DB"
	"github.com/sirupsen/logrus"
)

// PreviewResult contains the results of a dry-run preview
type PreviewResult struct {
	SourceTable       string
	DestTable         string
	ColumnMapping     map[string]string
	TotalRows         int
	EstimatedDuration time.Duration
	SampleData        []map[string]interface{}
	Warnings          []string
	SQLPreview        string
}

// MigrationPreviewer handles dry-run previews of migrations
type MigrationPreviewer struct {
	logger *logrus.Logger
}

// NewMigrationPreviewer creates a new migration previewer
func NewMigrationPreviewer() *MigrationPreviewer {
	return &MigrationPreviewer{
		logger: logrus.New(),
	}
}

// GeneratePreview generates a comprehensive preview of what the migration would do
func (p *MigrationPreviewer) GeneratePreview(
	sourceDB DB.DBInterface,
	destDB DB.DBInterface,
	sourceTable, destTable string,
	columnMapping map[string]string,
	sourceDBType, destDBType string,
) (*PreviewResult, error) {
	
	logrus.WithFields(logrus.Fields{
		"sourceTable": sourceTable,
		"destTable":   destTable,
		"mappings":    len(columnMapping),
	}).Info("Generating migration preview")

	result := &PreviewResult{
		SourceTable:   sourceTable,
		DestTable:     destTable,
		ColumnMapping: columnMapping,
		Warnings:      []string{},
		SampleData:    []map[string]interface{}{},
	}

	// Step 1: Count total rows
	if err := p.countTotalRows(sourceDB, sourceTable, result); err != nil {
		return nil, fmt.Errorf("failed to count rows: %v", err)
	}

	// Step 2: Estimate migration duration
	p.estimateDuration(result)

	// Step 3: Sample data preview
	if err := p.sampleData(sourceDB, sourceTable, columnMapping, sourceDBType, result); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to sample data: %v", err))
	}

	// Step 4: Generate SQL preview
	if err := p.generateSQLPreview(sourceDB, sourceTable, destTable, columnMapping, sourceDBType, destDBType, result); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to generate SQL preview: %v", err))
	}

	// Step 5: Identify potential issues
	p.identifyPotentialIssues(result)

	logrus.WithFields(logrus.Fields{
		"totalRows":      result.TotalRows,
		"warningCount":   len(result.Warnings),
		"sampleCount":    len(result.SampleData),
	}).Info("Migration preview generated")

	return result, nil
}

// countTotalRows counts the total number of rows to be migrated
func (p *MigrationPreviewer) countTotalRows(sourceDB DB.DBInterface, sourceTable string, result *PreviewResult) error {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", sourceTable)
	
	rows, err := sourceDB.RawQuery(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&result.TotalRows); err != nil {
			return err
		}
	}

	return nil
}

// estimateDuration estimates how long the migration might take
func (p *MigrationPreviewer) estimateDuration(result *PreviewResult) {
	// Base estimation: assume 1000 rows per second processing speed
	// This is a rough estimate and can be refined based on actual benchmarks
	rowsPerSecond := 1000.0
	
	if result.TotalRows == 0 {
		result.EstimatedDuration = 0
		return
	}
	
	// Adjust based on column count (more columns = slower processing)
	columnFactor := float64(len(result.ColumnMapping)) / 5.0 // Base of 5 columns
	if columnFactor < 0.5 {
		columnFactor = 0.5
	}
	
	adjustedSpeed := rowsPerSecond / columnFactor
	estimatedSeconds := float64(result.TotalRows) / adjustedSpeed
	
	// Add overhead for transaction management, validation, etc.
	estimatedSeconds *= 1.2
	
	result.EstimatedDuration = time.Duration(estimatedSeconds) * time.Second
}

// sampleData retrieves a sample of the data to be migrated
func (p *MigrationPreviewer) sampleData(
	sourceDB DB.DBInterface,
	sourceTable string,
	columnMapping map[string]string,
	sourceDBType string,
	result *PreviewResult,
) error {
	var sourceColumns []string
	for sourceCol := range columnMapping {
		sourceColumns = append(sourceColumns, sourceCol)
	}

	// Limit sample to first 10 rows
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(sourceColumns, ", "), sourceTable)
	if sourceDBType == "mysql" {
		query += " LIMIT 10"
	} else if sourceDBType == "oracle" {
		query = fmt.Sprintf("SELECT * FROM (%s) WHERE ROWNUM <= 10", query)
	}

	rows, err := sourceDB.RawQuery(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		values := make([]interface{}, len(sourceColumns))
		valuePtrs := make([]interface{}, len(sourceColumns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue // Skip problematic rows in preview
		}

		// Convert to map with column names
		rowData := make(map[string]interface{})
		for i, col := range sourceColumns {
			// Format value based on type
			if values[i] == nil {
				rowData[col] = nil
			} else {
				// Convert bytes to string for display
				if bytes, ok := values[i].([]byte); ok {
					rowData[col] = string(bytes)
				} else {
					rowData[col] = values[i]
				}
			}
		}

		result.SampleData = append(result.SampleData, rowData)
	}

	return nil
}

// generateSQLPreview generates a preview of the SQL statements that would be executed
func (p *MigrationPreviewer) generateSQLPreview(
	sourceDB DB.DBInterface,
	sourceTable, destTable string,
	columnMapping map[string]string,
	sourceDBType, destDBType string,
	result *PreviewResult,
) error {
	var sqlPreview strings.Builder
	
	sqlPreview.WriteString("-- Migration Preview: SQL Statements\n")
	sqlPreview.WriteString(fmt.Sprintf("-- Source: %s.%s\n", sourceDBType, sourceTable))
	sqlPreview.WriteString(fmt.Sprintf("-- Destination: %s.%s\n", destDBType, destTable))
	sqlPreview.WriteString(fmt.Sprintf("-- Estimated rows: %d\n\n", result.TotalRows))

	// Show the basic INSERT template
	var sourceColumns []string
	var destColumns []string
	for srcCol, destCol := range columnMapping {
		sourceColumns = append(sourceColumns, srcCol)
		destColumns = append(destColumns, destCol)
	}

	// Generate sample INSERT statement
	sqlPreview.WriteString("-- Sample INSERT statement template:\n")
	placeholders := make([]string, len(destColumns))
	for i := range placeholders {
		if destDBType == "oracle" {
			placeholders[i] = fmt.Sprintf(":v%d", i+1)
		} else {
			placeholders[i] = "?"
		}
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);",
		destTable,
		strings.Join(destColumns, ", "),
		strings.Join(placeholders, ", "))
	sqlPreview.WriteString(insertSQL)
	sqlPreview.WriteString("\n\n")

	// Show transaction structure
	sqlPreview.WriteString("-- Transaction structure:\n")
	sqlPreview.WriteString("BEGIN TRANSACTION;\n")
	sqlPreview.WriteString("-- Batch INSERT statements (1000 rows per batch)\n")
	sqlPreview.WriteString("-- ... INSERT statements ...\n")
	sqlPreview.WriteString("COMMIT;\n\n")

	// Column mapping info
	sqlPreview.WriteString("-- Column Mappings:\n")
	for srcCol, destCol := range columnMapping {
		sqlPreview.WriteString(fmt.Sprintf("-- %s → %s\n", srcCol, destCol))
	}

	result.SQLPreview = sqlPreview.String()
	return nil
}

// identifyPotentialIssues analyzes the migration setup for potential problems
func (p *MigrationPreviewer) identifyPotentialIssues(result *PreviewResult) {
	// Check for very large migrations
	if result.TotalRows > 1000000 {
		result.Warnings = append(result.Warnings, 
			fmt.Sprintf("Large migration detected (%d rows). Consider running during off-peak hours.", result.TotalRows))
	}

	// Check for long estimated duration
	if result.EstimatedDuration > 30*time.Minute {
		result.Warnings = append(result.Warnings, 
			fmt.Sprintf("Migration estimated to take %v. Consider breaking into smaller chunks.", result.EstimatedDuration))
	}

	// Check for many columns
	if len(result.ColumnMapping) > 20 {
		result.Warnings = append(result.Warnings, 
			fmt.Sprintf("Many columns mapped (%d). Verify all mappings are correct.", len(result.ColumnMapping)))
	}

	// Check sample data for potential issues
	for i, row := range result.SampleData {
		for col, value := range row {
			if value == nil {
				continue
			}
			
			// Check for very long strings
			if str, ok := value.(string); ok && len(str) > 1000 {
				result.Warnings = append(result.Warnings, 
					fmt.Sprintf("Long text value detected in column '%s' (sample row %d). May affect performance.", col, i+1))
				break // Don't spam warnings
			}
		}
	}

	// Check for empty tables
	if result.TotalRows == 0 {
		result.Warnings = append(result.Warnings, "Source table appears to be empty. Migration will have no effect.")
	}
}

// FormatPreviewForDisplay formats the preview result for console display
func (p *MigrationPreviewer) FormatPreviewForDisplay(result *PreviewResult) string {
	var output strings.Builder
	
	output.WriteString("╔════════════════════════════════════════════════════════════════════════╗\n")
	output.WriteString("║                           MIGRATION PREVIEW                           ║\n")
	output.WriteString("╠════════════════════════════════════════════════════════════════════════╣\n")
	
	// Basic info
	output.WriteString(fmt.Sprintf("║ Source Table: %-56s ║\n", result.SourceTable))
	output.WriteString(fmt.Sprintf("║ Destination Table: %-51s ║\n", result.DestTable))
	output.WriteString(fmt.Sprintf("║ Total Rows: %-58d ║\n", result.TotalRows))
	output.WriteString(fmt.Sprintf("║ Estimated Duration: %-50s ║\n", result.EstimatedDuration.String()))
	output.WriteString("║                                                                        ║\n")
	
	// Column mappings
	output.WriteString("║ Column Mappings:                                                       ║\n")
	for srcCol, destCol := range result.ColumnMapping {
		mapping := fmt.Sprintf("%s → %s", srcCol, destCol)
		output.WriteString(fmt.Sprintf("║   %-68s ║\n", mapping))
	}
	output.WriteString("║                                                                        ║\n")
	
	// Sample data
	if len(result.SampleData) > 0 {
		output.WriteString("║ Sample Data (first few rows):                                         ║\n")
		for i, row := range result.SampleData {
			if i >= 3 { // Limit display to first 3 rows
				break
			}
			output.WriteString(fmt.Sprintf("║ Row %d:                                                                ║\n", i+1))
			for col, value := range row {
				valueStr := fmt.Sprintf("%v", value)
				if len(valueStr) > 50 {
					valueStr = valueStr[:47] + "..."
				}
				output.WriteString(fmt.Sprintf("║   %s: %-50s ║\n", col, valueStr))
			}
		}
		output.WriteString("║                                                                        ║\n")
	}
	
	// Warnings
	if len(result.Warnings) > 0 {
		output.WriteString("║ ⚠️  WARNINGS:                                                          ║\n")
		for _, warning := range result.Warnings {
			// Wrap long warnings
			words := strings.Fields(warning)
			line := ""
			for _, word := range words {
				if len(line)+len(word)+1 > 66 {
					output.WriteString(fmt.Sprintf("║   %-68s ║\n", line))
					line = word
				} else {
					if line != "" {
						line += " "
					}
					line += word
				}
			}
			if line != "" {
				output.WriteString(fmt.Sprintf("║   %-68s ║\n", line))
			}
		}
		output.WriteString("║                                                                        ║\n")
	}
	
	output.WriteString("╚════════════════════════════════════════════════════════════════════════╝\n")
	
	return output.String()
}
