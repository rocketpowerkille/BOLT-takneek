package DB

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
	
	"github.com/sirupsen/logrus"
)

// MySQL utility functions
func FetchTables(db *gorm.DB) ([]string, error) {
	var tables []string
	err := db.Raw("SHOW TABLES").Scan(&tables).Error
	return tables, err
}

func FetchColumns(db *gorm.DB, table string) ([]string, error) {
	var cols []string
	err := db.Raw("SHOW COLUMNS FROM " + table).Scan(&cols).Error
	return cols, err
}

// Oracle utility functions
func FetchOracleTables(db *sql.DB) ([]string, error) {
	query := "SELECT table_name FROM user_tables ORDER BY table_name"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}
	return tables, rows.Err()
}

func FetchOracleColumns(db *sql.DB, table string) ([]string, error) {
	query := "SELECT column_name FROM user_tab_columns WHERE table_name = :1 ORDER BY column_id"
	rows, err := db.Query(query, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		columns = append(columns, columnName)
	}
	return columns, rows.Err()
}

// Generic functions that work with DBInterface
func FetchTablesGeneric(db DBInterface, dbType string) ([]string, error) {
	var query string
	switch dbType {
	case "mysql":
		query = "SHOW TABLES"
	case "oracle":
		query = "SELECT table_name FROM user_tables ORDER BY table_name"
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	rows, err := db.RawQuery(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}
	return tables, rows.Err()
}

func FetchColumnsGeneric(db DBInterface, table string, dbType string) ([]string, error) {
	var query string
	switch dbType {
	case "mysql":
		query = fmt.Sprintf("SHOW COLUMNS FROM %s", table)
	case "oracle":
		query = "SELECT column_name FROM user_tab_columns WHERE table_name = ? ORDER BY column_id"
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	var rows *sql.Rows
	var err error
	
	if dbType == "oracle" {
		rows, err = db.RawQuery(query, table)
	} else {
		rows, err = db.RawQuery(query)
	}
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var columnName string
		if dbType == "mysql" {
			// MySQL SHOW COLUMNS returns multiple fields, we only want the first one
			var dataType, null, key, defaultVal, extra string
			if err := rows.Scan(&columnName, &dataType, &null, &key, &defaultVal, &extra); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(&columnName); err != nil {
				return nil, err
			}
		}
		columns = append(columns, columnName)
	}
	return columns, rows.Err()
}

// Data migration functions with transaction support
func MigrateData(source DBInterface, dest DBInterface, sourceTable, destTable string, columnMapping map[string]string) error {
	return MigrateDataWithProgress(source, dest, sourceTable, destTable, columnMapping, nil)
}

// MigrateDataWithProgress performs migration with optional progress callback
func MigrateDataWithProgress(source DBInterface, dest DBInterface, sourceTable, destTable string, columnMapping map[string]string, progressCallback func(processed, total int)) error {
	return MigrateDataWithProgressAndDBType(source, dest, sourceTable, destTable, columnMapping, progressCallback, "mysql", "mysql")
}

// MigrateDataWithProgressAndDBType performs migration with database type awareness
func MigrateDataWithProgressAndDBType(source DBInterface, dest DBInterface, sourceTable, destTable string, columnMapping map[string]string, progressCallback func(processed, total int), sourceDBType, destDBType string) error {
	startTime := time.Now()
	logrus.WithFields(logrus.Fields{
		"sourceTable": sourceTable,
		"destTable":   destTable,
		"mappings":    len(columnMapping),
		"sourceType":  sourceDBType,
		"destType":    destDBType,
	}).Info("Starting data migration")

	if len(columnMapping) == 0 {
		return fmt.Errorf("no column mapping provided")
	}

	// Build the source query
	var sourceColumns []string
	var destColumns []string
	for srcCol, destCol := range columnMapping {
		sourceColumns = append(sourceColumns, srcCol)
		destColumns = append(destColumns, destCol)
	}

	logrus.WithFields(logrus.Fields{
		"sourceColumns": sourceColumns,
		"destColumns":   destColumns,
	}).Debug("Column mapping prepared")

	// Start transaction for destination
	logrus.Info("Beginning destination database transaction")
	destTx, err := dest.BeginTx()
	if err != nil {
		logrus.WithError(err).Error("Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err := destTx.Rollback(); err != nil {
			logrus.WithError(err).Debug("Transaction rollback completed (this is normal if commit succeeded)")
		}
	}()

	// First, get total count for progress tracking
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", sourceTable)
	logrus.WithField("query", countQuery).Debug("Counting source rows")
	
	countRows, err := source.RawQuery(countQuery)
	if err != nil {
		logrus.WithError(err).Error("Failed to count source rows")
		return fmt.Errorf("error counting source rows: %v", err)
	}
	defer countRows.Close()

	var totalRows int
	if countRows.Next() {
		if err := countRows.Scan(&totalRows); err != nil {
			logrus.WithError(err).Error("Failed to scan row count")
			return fmt.Errorf("error scanning count: %v", err)
		}
	}
	countRows.Close()

	logrus.WithField("totalRows", totalRows).Info("Source row count obtained")

	// Fetch data from source with pagination for large datasets
	const batchSize = 1000
	offset := 0
	processedRows := 0

	logrus.WithField("batchSize", batchSize).Info("Starting batch processing")

	for {
		batchStartTime := time.Now()
		
		// Build paginated query based on database type
		var sourceQuery string
		if sourceDBType == "oracle" {
			// Oracle uses ROWNUM or OFFSET/FETCH
			sourceQuery = fmt.Sprintf("SELECT %s FROM %s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", 
				strings.Join(sourceColumns, ", "), sourceTable, offset, batchSize)
		} else {
			// MySQL uses LIMIT/OFFSET
			sourceQuery = fmt.Sprintf("SELECT %s FROM %s LIMIT %d OFFSET %d", 
				strings.Join(sourceColumns, ", "), sourceTable, batchSize, offset)
		}
		
		logrus.WithFields(logrus.Fields{
			"offset":    offset,
			"batchSize": batchSize,
		}).Debug("Processing batch")

		rows, err := source.RawQuery(sourceQuery)
		if err != nil {
			logrus.WithError(err).WithField("query", sourceQuery).Error("Failed to query source table")
			return fmt.Errorf("error querying source table: %v", err)
		}

		// Prepare destination insert query
		placeholders := make([]string, len(destColumns))
		for i := range placeholders {
			if destDBType == "oracle" {
				placeholders[i] = fmt.Sprintf(":v%d", i+1)
			} else {
				placeholders[i] = "?"
			}
		}
		destQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", 
			destTable, 
			strings.Join(destColumns, ", "), 
			strings.Join(placeholders, ", "))

		// Process batch
		batchCount := 0
		batchValues := make([][]interface{}, 0, batchSize)

		for rows.Next() {
			// Create slice to hold values
			values := make([]interface{}, len(sourceColumns))
			valuePtrs := make([]interface{}, len(sourceColumns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			// Scan row values
			if err := rows.Scan(valuePtrs...); err != nil {
				rows.Close()
				logrus.WithError(err).WithField("rowNum", processedRows+batchCount+1).Error("Failed to scan row")
				return fmt.Errorf("error scanning row %d: %v", processedRows+batchCount+1, err)
			}

			batchValues = append(batchValues, values)
			batchCount++
		}
		rows.Close()

		// If no more rows, break
		if batchCount == 0 {
			logrus.Info("No more rows to process")
			break
		}

		// Insert batch into destination
		for _, values := range batchValues {
			if err := destTx.ExecQuery(destQuery, values...); err != nil {
				logrus.WithError(err).WithField("rowNum", processedRows+1).Error("Failed to insert row")
				return fmt.Errorf("error inserting row %d: %v", processedRows+1, err)
			}
			processedRows++

			// Call progress callback if provided
			if progressCallback != nil {
				progressCallback(processedRows, totalRows)
			}
		}

		batchDuration := time.Since(batchStartTime)
		logrus.WithFields(logrus.Fields{
			"batchCount":     batchCount,
			"processedRows":  processedRows,
			"totalRows":      totalRows,
			"batchDuration":  batchDuration,
			"rowsPerSecond":  float64(batchCount) / batchDuration.Seconds(),
		}).Info("Batch processed")

		// If we got fewer rows than batch size, we're done
		if batchCount < batchSize {
			break
		}

		offset += batchSize
	}

	// Commit transaction
	logrus.Info("Committing transaction")
	if err := destTx.Commit(); err != nil {
		logrus.WithError(err).Error("Failed to commit transaction")
		return fmt.Errorf("error committing transaction: %v", err)
	}

	totalDuration := time.Since(startTime)
	logrus.WithFields(logrus.Fields{
		"totalRows":      processedRows,
		"totalDuration":  totalDuration,
		"rowsPerSecond":  float64(processedRows) / totalDuration.Seconds(),
	}).Info("Migration completed successfully")

	return nil
}

// Generate SQL dump
func GenerateSQLDump(source DBInterface, sourceTable string, columnMapping map[string]string, destTable string) (string, error) {
	if len(columnMapping) == 0 {
		return "", fmt.Errorf("no column mapping provided")
	}

	var sourceColumns []string
	var destColumns []string
	for srcCol, destCol := range columnMapping {
		sourceColumns = append(sourceColumns, srcCol)
		destColumns = append(destColumns, destCol)
	}

	sourceQuery := fmt.Sprintf("SELECT %s FROM %s", strings.Join(sourceColumns, ", "), sourceTable)
	
	rows, err := source.RawQuery(sourceQuery)
	if err != nil {
		return "", fmt.Errorf("error querying source table: %v", err)
	}
	defer rows.Close()

	var sqlDump strings.Builder
	sqlDump.WriteString(fmt.Sprintf("-- Data migration from %s to %s\n", sourceTable, destTable))
	sqlDump.WriteString(fmt.Sprintf("-- Generated column mapping: %v\n\n", columnMapping))

	for rows.Next() {
		values := make([]interface{}, len(sourceColumns))
		valuePtrs := make([]interface{}, len(sourceColumns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return "", fmt.Errorf("error scanning row: %v", err)
		}

		// Convert values to strings for SQL
		valueStrings := make([]string, len(values))
		for i, v := range values {
			if v == nil {
				valueStrings[i] = "NULL"
			} else {
				valueStrings[i] = fmt.Sprintf("'%v'", v)
			}
		}

		sqlDump.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n",
			destTable,
			strings.Join(destColumns, ", "),
			strings.Join(valueStrings, ", ")))
	}

	return sqlDump.String(), rows.Err()
}
