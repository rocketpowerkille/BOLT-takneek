package validation

import (
	"fmt"
	"strings"
	"database/sql"
	
	"github.com/pclubiitk/dbcli/DB"
	"github.com/pclubiitk/dbcli/config"
	"github.com/sirupsen/logrus"
)

// ColumnInfo represents column metadata for validation
type ColumnInfo struct {
	Name         string
	DataType     string
	IsNullable   bool
	MaxLength    *int
	DefaultValue *string
	IsPrimaryKey bool
	IsForeignKey bool
	FKTable      *string
	FKColumn     *string
}

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	IsValid   bool
	Errors    []string
	Warnings  []string
	RowCount  int
	Issues    []ValidationIssue
}

// ValidationIssue represents a specific validation problem
type ValidationIssue struct {
	Type        string // "data_type", "null_constraint", "foreign_key", "length", etc.
	RowNumber   int
	ColumnName  string
	Value       interface{}
	Description string
	Severity    string // "error", "warning"
}

// DataValidator handles data validation before migration
type DataValidator struct {
	config *config.Config
	logger *logrus.Logger
}

// NewDataValidator creates a new data validator
func NewDataValidator(cfg *config.Config) *DataValidator {
	return &DataValidator{
		config: cfg,
		logger: logrus.New(),
	}
}

// ValidateBeforeMigration performs comprehensive validation before migration
func (v *DataValidator) ValidateBeforeMigration(
	sourceDB DB.DBInterface,
	destDB DB.DBInterface,
	sourceTable, destTable string,
	columnMapping map[string]string,
	sourceDBType, destDBType string,
) (*ValidationResult, error) {
	
	logrus.WithFields(logrus.Fields{
		"sourceTable": sourceTable,
		"destTable":   destTable,
		"mappings":    len(columnMapping),
	}).Info("Starting data validation")

	result := &ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
		Issues:   []ValidationIssue{},
	}

	// Step 1: Validate column mappings exist
	if err := v.validateColumnMappings(sourceDB, destDB, sourceTable, destTable, columnMapping, sourceDBType, destDBType, result); err != nil {
		return result, err
	}

	// Step 2: Validate data types compatibility
	if v.config.Migration.ValidationRules.CheckDataTypes {
		if err := v.validateDataTypes(sourceDB, destDB, sourceTable, destTable, columnMapping, sourceDBType, destDBType, result); err != nil {
			return result, err
		}
	}

	// Step 3: Validate NULL constraints
	if v.config.Migration.ValidationRules.CheckNullConstraints {
		if err := v.validateNullConstraints(sourceDB, destDB, sourceTable, destTable, columnMapping, sourceDBType, destDBType, result); err != nil {
			return result, err
		}
	}

	// Step 4: Validate foreign key constraints
	if v.config.Migration.ValidationRules.CheckForeignKeys {
		if err := v.validateForeignKeys(sourceDB, destDB, sourceTable, destTable, columnMapping, sourceDBType, destDBType, result); err != nil {
			return result, err
		}
	}

	// Step 5: Sample data validation
	if err := v.validateSampleData(sourceDB, destDB, sourceTable, destTable, columnMapping, sourceDBType, destDBType, result); err != nil {
		return result, err
	}

	// Determine overall validity
	if len(result.Errors) > 0 {
		result.IsValid = false
	}

	if v.config.Migration.ValidationRules.StrictMode && len(result.Warnings) > 0 {
		result.IsValid = false
		// Convert warnings to errors in strict mode
		for _, warning := range result.Warnings {
			result.Errors = append(result.Errors, fmt.Sprintf("STRICT MODE: %s", warning))
		}
	}

	logrus.WithFields(logrus.Fields{
		"isValid":     result.IsValid,
		"errorCount":  len(result.Errors),
		"warningCount": len(result.Warnings),
		"issueCount":  len(result.Issues),
	}).Info("Data validation completed")

	return result, nil
}

// validateColumnMappings checks if all mapped columns exist
func (v *DataValidator) validateColumnMappings(
	sourceDB, destDB DB.DBInterface,
	sourceTable, destTable string,
	columnMapping map[string]string,
	sourceDBType, destDBType string,
	result *ValidationResult,
) error {
	// Get source columns
	sourceColumns, err := DB.FetchColumnsGeneric(sourceDB, sourceTable, sourceDBType)
	if err != nil {
		return fmt.Errorf("failed to fetch source columns: %v", err)
	}

	// Get destination columns
	destColumns, err := DB.FetchColumnsGeneric(destDB, destTable, destDBType)
	if err != nil {
		return fmt.Errorf("failed to fetch destination columns: %v", err)
	}

	// Convert to maps for faster lookup
	sourceColMap := make(map[string]bool)
	for _, col := range sourceColumns {
		sourceColMap[strings.ToLower(col)] = true
	}

	destColMap := make(map[string]bool)
	for _, col := range destColumns {
		destColMap[strings.ToLower(col)] = true
	}

	// Validate mappings
	for sourceCol, destCol := range columnMapping {
		if !sourceColMap[strings.ToLower(sourceCol)] {
			result.Errors = append(result.Errors, fmt.Sprintf("Source column '%s' does not exist in table '%s'", sourceCol, sourceTable))
			result.Issues = append(result.Issues, ValidationIssue{
				Type:        "missing_column",
				ColumnName:  sourceCol,
				Description: fmt.Sprintf("Source column '%s' not found", sourceCol),
				Severity:    "error",
			})
		}

		if !destColMap[strings.ToLower(destCol)] {
			result.Errors = append(result.Errors, fmt.Sprintf("Destination column '%s' does not exist in table '%s'", destCol, destTable))
			result.Issues = append(result.Issues, ValidationIssue{
				Type:        "missing_column",
				ColumnName:  destCol,
				Description: fmt.Sprintf("Destination column '%s' not found", destCol),
				Severity:    "error",
			})
		}
	}

	return nil
}

// validateDataTypes checks data type compatibility
func (v *DataValidator) validateDataTypes(
	sourceDB, destDB DB.DBInterface,
	sourceTable, destTable string,
	columnMapping map[string]string,
	sourceDBType, destDBType string,
	result *ValidationResult,
) error {
	// Get detailed column information
	sourceColInfo, err := v.getColumnInfo(sourceDB, sourceTable, sourceDBType)
	if err != nil {
		return fmt.Errorf("failed to get source column info: %v", err)
	}

	destColInfo, err := v.getColumnInfo(destDB, destTable, destDBType)
	if err != nil {
		return fmt.Errorf("failed to get destination column info: %v", err)
	}

	// Check data type compatibility for each mapping
	for sourceCol, destCol := range columnMapping {
		sourceInfo, sourceExists := sourceColInfo[strings.ToLower(sourceCol)]
		destInfo, destExists := destColInfo[strings.ToLower(destCol)]

		if !sourceExists || !destExists {
			continue // Already handled in column mapping validation
		}

		// Check data type compatibility
		compatible, warning := v.areDataTypesCompatible(sourceInfo.DataType, destInfo.DataType, sourceDBType, destDBType)
		if !compatible {
			result.Errors = append(result.Errors, fmt.Sprintf("Incompatible data types: %s.%s (%s) -> %s.%s (%s)", 
				sourceTable, sourceCol, sourceInfo.DataType, destTable, destCol, destInfo.DataType))
			result.Issues = append(result.Issues, ValidationIssue{
				Type:        "data_type_incompatible",
				ColumnName:  fmt.Sprintf("%s -> %s", sourceCol, destCol),
				Description: fmt.Sprintf("Data type mismatch: %s -> %s", sourceInfo.DataType, destInfo.DataType),
				Severity:    "error",
			})
		} else if warning != "" {
			result.Warnings = append(result.Warnings, warning)
			result.Issues = append(result.Issues, ValidationIssue{
				Type:        "data_type_warning",
				ColumnName:  fmt.Sprintf("%s -> %s", sourceCol, destCol),
				Description: warning,
				Severity:    "warning",
			})
		}

		// Check length constraints
		if sourceInfo.MaxLength != nil && destInfo.MaxLength != nil {
			if *sourceInfo.MaxLength > *destInfo.MaxLength {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Source column %s.%s has max length %d, destination %s.%s has max length %d - data may be truncated",
					sourceTable, sourceCol, *sourceInfo.MaxLength, destTable, destCol, *destInfo.MaxLength))
				result.Issues = append(result.Issues, ValidationIssue{
					Type:        "length_mismatch",
					ColumnName:  fmt.Sprintf("%s -> %s", sourceCol, destCol),
					Description: fmt.Sprintf("Potential data truncation: %d -> %d", *sourceInfo.MaxLength, *destInfo.MaxLength),
					Severity:    "warning",
				})
			}
		}
	}

	return nil
}

// validateNullConstraints checks NULL constraint compatibility
func (v *DataValidator) validateNullConstraints(
	sourceDB, destDB DB.DBInterface,
	sourceTable, destTable string,
	columnMapping map[string]string,
	sourceDBType, destDBType string,
	result *ValidationResult,
) error {
	// Get column information
	sourceColInfo, err := v.getColumnInfo(sourceDB, sourceTable, sourceDBType)
	if err != nil {
		return fmt.Errorf("failed to get source column info: %v", err)
	}

	destColInfo, err := v.getColumnInfo(destDB, destTable, destDBType)
	if err != nil {
		return fmt.Errorf("failed to get destination column info: %v", err)
	}

	// Check NULL constraints
	for sourceCol, destCol := range columnMapping {
		sourceInfo, sourceExists := sourceColInfo[strings.ToLower(sourceCol)]
		destInfo, destExists := destColInfo[strings.ToLower(destCol)]

		if !sourceExists || !destExists {
			continue
		}

		// If source allows NULLs but destination doesn't, check for NULL values
		if sourceInfo.IsNullable && !destInfo.IsNullable {
			nullCount, err := v.countNullValues(sourceDB, sourceTable, sourceCol, sourceDBType)
			if err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Could not check NULL values in %s.%s: %v", sourceTable, sourceCol, err))
				continue
			}

			if nullCount > 0 {
				result.Errors = append(result.Errors, fmt.Sprintf("Column %s.%s contains %d NULL values, but destination column %s.%s does not allow NULLs",
					sourceTable, sourceCol, nullCount, destTable, destCol))
				result.Issues = append(result.Issues, ValidationIssue{
					Type:        "null_constraint_violation",
					ColumnName:  fmt.Sprintf("%s -> %s", sourceCol, destCol),
					Description: fmt.Sprintf("%d NULL values would violate NOT NULL constraint", nullCount),
					Severity:    "error",
				})
			}
		}
	}

	return nil
}

// validateForeignKeys checks foreign key constraint compatibility
func (v *DataValidator) validateForeignKeys(
	sourceDB, destDB DB.DBInterface,
	sourceTable, destTable string,
	columnMapping map[string]string,
	sourceDBType, destDBType string,
	result *ValidationResult,
) error {
	// This is a simplified foreign key validation
	// In a production system, you'd want more sophisticated FK checking
	result.Warnings = append(result.Warnings, "Foreign key validation is not fully implemented - manual verification recommended")
	
	return nil
}

// validateSampleData performs validation on a sample of the actual data
func (v *DataValidator) validateSampleData(
	sourceDB, destDB DB.DBInterface,
	sourceTable, destTable string,
	columnMapping map[string]string,
	sourceDBType, destDBType string,
	result *ValidationResult,
) error {
	// Sample first 100 rows for validation
	sampleSize := 100
	
	var sourceColumns []string
	for sourceCol := range columnMapping {
		sourceColumns = append(sourceColumns, sourceCol)
	}

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(sourceColumns, ", "), sourceTable)
	if sourceDBType == "mysql" {
		query += fmt.Sprintf(" LIMIT %d", sampleSize)
	} else if sourceDBType == "oracle" {
		query = fmt.Sprintf("SELECT * FROM (%s) WHERE ROWNUM <= %d", query, sampleSize)
	}

	rows, err := sourceDB.RawQuery(query)
	if err != nil {
		return fmt.Errorf("failed to sample data: %v", err)
	}
	defer rows.Close()

	rowCount := 0
	for rows.Next() {
		rowCount++
		
		values := make([]interface{}, len(sourceColumns))
		valuePtrs := make([]interface{}, len(sourceColumns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to scan row %d: %v", rowCount, err))
			continue
		}

		// Validate each value
		for i, sourceCol := range sourceColumns {
			value := values[i]
			if err := v.validateValue(value, sourceCol, columnMapping[sourceCol], rowCount, result); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Row %d, column %s: %v", rowCount, sourceCol, err))
			}
		}
	}

	result.RowCount = rowCount
	logrus.WithField("sampleSize", rowCount).Info("Sample data validation completed")

	return nil
}

// getColumnInfo retrieves detailed column information
func (v *DataValidator) getColumnInfo(db DB.DBInterface, table, dbType string) (map[string]ColumnInfo, error) {
	colInfo := make(map[string]ColumnInfo)

	var query string
	switch dbType {
	case "mysql":
		query = fmt.Sprintf(`
			SELECT 
				COLUMN_NAME,
				DATA_TYPE,
				IS_NULLABLE,
				CHARACTER_MAXIMUM_LENGTH,
				COLUMN_DEFAULT,
				COLUMN_KEY
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_NAME = '%s'
		`, table)
	case "oracle":
		query = fmt.Sprintf(`
			SELECT 
				COLUMN_NAME,
				DATA_TYPE,
				NULLABLE,
				CHAR_LENGTH,
				DATA_DEFAULT,
				''
			FROM USER_TAB_COLUMNS 
			WHERE TABLE_NAME = '%s'
		`, strings.ToUpper(table))
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	rows, err := db.RawQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get column info: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var colName, dataType, nullable, columnKey string
		var maxLength sql.NullInt64
		var defaultValue sql.NullString

		if err := rows.Scan(&colName, &dataType, &nullable, &maxLength, &defaultValue, &columnKey); err != nil {
			continue
		}

		info := ColumnInfo{
			Name:         colName,
			DataType:     dataType,
			IsNullable:   nullable == "YES" || nullable == "Y",
			IsPrimaryKey: columnKey == "PRI",
		}

		if maxLength.Valid {
			length := int(maxLength.Int64)
			info.MaxLength = &length
		}

		if defaultValue.Valid {
			info.DefaultValue = &defaultValue.String
		}

		colInfo[strings.ToLower(colName)] = info
	}

	return colInfo, nil
}

// areDataTypesCompatible checks if two data types are compatible
func (v *DataValidator) areDataTypesCompatible(sourceType, destType, sourceDBType, destDBType string) (bool, string) {
	sourceType = strings.ToLower(sourceType)
	destType = strings.ToLower(destType)

	// Exact match
	if sourceType == destType {
		return true, ""
	}

	// Common compatible types
	compatibleTypes := map[string][]string{
		"varchar":  {"varchar", "text", "char", "nvarchar"},
		"char":     {"varchar", "text", "char", "nvarchar"},
		"text":     {"varchar", "text", "longtext", "mediumtext"},
		"int":      {"int", "integer", "bigint", "smallint", "number"},
		"integer":  {"int", "integer", "bigint", "smallint", "number"},
		"bigint":   {"int", "integer", "bigint", "number"},
		"decimal":  {"decimal", "numeric", "number", "float", "double"},
		"numeric":  {"decimal", "numeric", "number", "float", "double"},
		"number":   {"decimal", "numeric", "number", "float", "double", "int", "integer", "bigint"},
		"date":     {"date", "datetime", "timestamp"},
		"datetime": {"date", "datetime", "timestamp"},
		"timestamp": {"date", "datetime", "timestamp"},
	}

	// Check if destination type is in compatible list for source type
	if compatibleList, exists := compatibleTypes[sourceType]; exists {
		for _, compatible := range compatibleList {
			if strings.Contains(destType, compatible) {
				if sourceType != destType {
					return true, fmt.Sprintf("Data type conversion from %s to %s may require attention", sourceType, destType)
				}
				return true, ""
			}
		}
	}

	return false, ""
}

// countNullValues counts NULL values in a column
func (v *DataValidator) countNullValues(db DB.DBInterface, table, column, dbType string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s IS NULL", table, column)
	
	rows, err := db.RawQuery(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
	}

	return count, nil
}

// validateValue validates a single value
func (v *DataValidator) validateValue(value interface{}, sourceCol, destCol string, rowNumber int, result *ValidationResult) error {
	// Basic value validation
	if value == nil {
		return nil // NULL values are handled separately
	}

	// Convert to string for basic checks
	valueStr := fmt.Sprintf("%v", value)
	
	// Check for potentially problematic characters
	if strings.Contains(valueStr, "\x00") {
		result.Issues = append(result.Issues, ValidationIssue{
			Type:        "invalid_character",
			RowNumber:   rowNumber,
			ColumnName:  sourceCol,
			Value:       value,
			Description: "Contains null byte character",
			Severity:    "warning",
		})
	}

	// Check for very long strings that might cause issues
	if len(valueStr) > 65535 {
		result.Issues = append(result.Issues, ValidationIssue{
			Type:        "long_value",
			RowNumber:   rowNumber,
			ColumnName:  sourceCol,
			Value:       fmt.Sprintf("Length: %d", len(valueStr)),
			Description: "Very long value may cause performance issues",
			Severity:    "warning",
		})
	}

	return nil
}
