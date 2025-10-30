package logging

import (
	"io"
	"os"
	"path/filepath"
	
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/sirupsen/logrus"
	"github.com/pclubiitk/dbcli/config"
)

// Logger represents an enhanced logger with file output and rotation
type Logger struct {
	*logrus.Logger
	config *config.Config
	fileWriter *lumberjack.Logger
}

// NewLogger creates a new enhanced logger based on configuration
func NewLogger(cfg *config.Config) (*Logger, error) {
	logger := logrus.New()
	
	// Set log level
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	
	enhancedLogger := &Logger{
		Logger: logger,
		config: cfg,
	}
	
	// Configure output based on settings
	if err := enhancedLogger.configureOutput(); err != nil {
		return nil, err
	}
	
	// Configure formatter
	enhancedLogger.configureFormatter()
	
	return enhancedLogger, nil
}

// configureOutput sets up console and/or file output
func (l *Logger) configureOutput() error {
	var outputs []io.Writer
	
	// Add console output if enabled
	if l.config.Logging.EnableConsole {
		outputs = append(outputs, os.Stdout)
	}
	
	// Add file output if enabled
	if l.config.Logging.EnableFileOutput {
		// Ensure log directory exists
		logDir := filepath.Dir(l.config.Logging.LogFilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}
		
		// Configure log rotation
		l.fileWriter = &lumberjack.Logger{
			Filename:   l.config.Logging.LogFilePath,
			MaxSize:    l.config.Logging.MaxFileSize, // MB
			MaxBackups: l.config.Logging.MaxBackups,
			MaxAge:     30, // days
			Compress:   true,
		}
		
		outputs = append(outputs, l.fileWriter)
	}
	
	// Set multi-writer if multiple outputs
	if len(outputs) > 1 {
		l.Logger.SetOutput(io.MultiWriter(outputs...))
	} else if len(outputs) == 1 {
		l.Logger.SetOutput(outputs[0])
	}
	
	return nil
}

// configureFormatter sets up the log formatter
func (l *Logger) configureFormatter() {
	if l.config.Logging.EnableJSON {
		// JSON formatter for structured logging
		l.Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		// Text formatter for human-readable logs
		l.Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     l.config.UI.EnableColors && l.config.Logging.EnableConsole,
		})
	}
}

// Close closes the file writer if it exists
func (l *Logger) Close() error {
	if l.fileWriter != nil {
		return l.fileWriter.Close()
	}
	return nil
}

// Migration-specific logging methods

// LogMigrationStart logs the start of a migration
func (l *Logger) LogMigrationStart(sourceTable, destTable string, columnCount int) {
	l.WithFields(logrus.Fields{
		"event":        "migration_start",
		"source_table": sourceTable,
		"dest_table":   destTable,
		"column_count": columnCount,
		"component":    "migration",
	}).Info("Migration started")
}

// LogMigrationProgress logs migration progress
func (l *Logger) LogMigrationProgress(processedRows, totalRows int, batchNum int) {
	percentage := float64(processedRows) / float64(totalRows) * 100
	l.WithFields(logrus.Fields{
		"event":          "migration_progress",
		"processed_rows": processedRows,
		"total_rows":     totalRows,
		"batch_number":   batchNum,
		"percentage":     percentage,
		"component":      "migration",
	}).Info("Migration progress")
}

// LogMigrationComplete logs successful migration completion
func (l *Logger) LogMigrationComplete(sourceTable, destTable string, totalRows int, duration string) {
	l.WithFields(logrus.Fields{
		"event":        "migration_complete",
		"source_table": sourceTable,
		"dest_table":   destTable,
		"total_rows":   totalRows,
		"duration":     duration,
		"component":    "migration",
	}).Info("Migration completed successfully")
}

// LogMigrationError logs migration errors
func (l *Logger) LogMigrationError(sourceTable, destTable string, err error, context map[string]interface{}) {
	fields := logrus.Fields{
		"event":        "migration_error",
		"source_table": sourceTable,
		"dest_table":   destTable,
		"error":        err.Error(),
		"component":    "migration",
	}
	
	// Add context fields
	for k, v := range context {
		fields[k] = v
	}
	
	l.WithFields(fields).Error("Migration failed")
}

// LogConnectionEvent logs database connection events
func (l *Logger) LogConnectionEvent(dbType, event string, details map[string]interface{}) {
	fields := logrus.Fields{
		"event":     "connection_" + event,
		"db_type":   dbType,
		"component": "database",
	}
	
	for k, v := range details {
		fields[k] = v
	}
	
	l.WithFields(fields).Info("Database connection event")
}

// LogValidationResult logs data validation results
func (l *Logger) LogValidationResult(isValid bool, errorCount, warningCount int) {
	l.WithFields(logrus.Fields{
		"event":         "validation_complete",
		"is_valid":      isValid,
		"error_count":   errorCount,
		"warning_count": warningCount,
		"component":     "validation",
	}).Info("Data validation completed")
}

// LogPerformanceMetrics logs performance metrics
func (l *Logger) LogPerformanceMetrics(metrics map[string]interface{}) {
	fields := logrus.Fields{
		"event":     "performance_metrics",
		"component": "performance",
	}
	
	for k, v := range metrics {
		fields[k] = v
	}
	
	l.WithFields(fields).Info("Performance metrics")
}

// LogCheckpointSaved logs when a checkpoint is saved
func (l *Logger) LogCheckpointSaved(checkpointID string, processedRows, totalRows int) {
	l.WithFields(logrus.Fields{
		"event":          "checkpoint_saved",
		"checkpoint_id":  checkpointID,
		"processed_rows": processedRows,
		"total_rows":     totalRows,
		"component":      "checkpoint",
	}).Info("Migration checkpoint saved")
}

// LogCheckpointLoaded logs when a checkpoint is loaded
func (l *Logger) LogCheckpointLoaded(checkpointID string, resumePoint int) {
	l.WithFields(logrus.Fields{
		"event":        "checkpoint_loaded",
		"checkpoint_id": checkpointID,
		"resume_point": resumePoint,
		"component":    "checkpoint",
	}).Info("Migration checkpoint loaded")
}
