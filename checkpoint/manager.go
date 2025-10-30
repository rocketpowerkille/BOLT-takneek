package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	
	"github.com/sirupsen/logrus"
)

// MigrationCheckpoint represents a saved migration state
type MigrationCheckpoint struct {
	ID              string                 `json:"id"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	SourceTable     string                `json:"source_table"`
	DestTable       string                `json:"dest_table"`
	ColumnMapping   map[string]string     `json:"column_mapping"`
	ProcessedRows   int                   `json:"processed_rows"`
	TotalRows       int                   `json:"total_rows"`
	BatchSize       int                   `json:"batch_size"`
	CurrentBatch    int                   `json:"current_batch"`
	Status          string                `json:"status"` // "running", "paused", "completed", "failed"
	LastError       string                `json:"last_error,omitempty"`
	SourceDBType    string                `json:"source_db_type"`
	DestDBType      string                `json:"dest_db_type"`
	MigrationConfig map[string]interface{} `json:"migration_config"`
}

// CheckpointManager handles saving and loading migration checkpoints
type CheckpointManager struct {
	checkpointDir string
	logger        *logrus.Logger
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(checkpointDir string) *CheckpointManager {
	if checkpointDir == "" {
		checkpointDir = "./checkpoints"
	}
	
	// Ensure checkpoint directory exists
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		logrus.WithError(err).Error("Failed to create checkpoint directory")
	}
	
	return &CheckpointManager{
		checkpointDir: checkpointDir,
		logger:        logrus.New(),
	}
}

// SaveCheckpoint saves the current migration state
func (cm *CheckpointManager) SaveCheckpoint(checkpoint *MigrationCheckpoint) error {
	checkpoint.UpdatedAt = time.Now()
	
	if checkpoint.ID == "" {
		checkpoint.ID = fmt.Sprintf("migration_%d", time.Now().Unix())
		checkpoint.CreatedAt = checkpoint.UpdatedAt
	}
	
	filename := fmt.Sprintf("%s.json", checkpoint.ID)
	filepath := filepath.Join(cm.checkpointDir, filename)
	
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %v", err)
	}
	
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write checkpoint file: %v", err)
	}
	
	logrus.WithFields(logrus.Fields{
		"checkpointID": checkpoint.ID,
		"processedRows": checkpoint.ProcessedRows,
		"totalRows": checkpoint.TotalRows,
		"status": checkpoint.Status,
	}).Info("Checkpoint saved")
	
	return nil
}

// LoadCheckpoint loads a migration checkpoint by ID
func (cm *CheckpointManager) LoadCheckpoint(checkpointID string) (*MigrationCheckpoint, error) {
	filename := fmt.Sprintf("%s.json", checkpointID)
	filepath := filepath.Join(cm.checkpointDir, filename)
	
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %v", err)
	}
	
	var checkpoint MigrationCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %v", err)
	}
	
	logrus.WithField("checkpointID", checkpointID).Info("Checkpoint loaded")
	return &checkpoint, nil
}

// ListCheckpoints returns a list of available checkpoints
func (cm *CheckpointManager) ListCheckpoints() ([]*MigrationCheckpoint, error) {
	files, err := os.ReadDir(cm.checkpointDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint directory: %v", err)
	}
	
	var checkpoints []*MigrationCheckpoint
	
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}
		
		checkpointID := file.Name()[:len(file.Name())-5] // Remove .json extension
		checkpoint, err := cm.LoadCheckpoint(checkpointID)
		if err != nil {
			logrus.WithError(err).WithField("file", file.Name()).Warn("Failed to load checkpoint")
			continue
		}
		
		checkpoints = append(checkpoints, checkpoint)
	}
	
	return checkpoints, nil
}

// DeleteCheckpoint removes a checkpoint file
func (cm *CheckpointManager) DeleteCheckpoint(checkpointID string) error {
	filename := fmt.Sprintf("%s.json", checkpointID)
	filepath := filepath.Join(cm.checkpointDir, filename)
	
	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete checkpoint: %v", err)
	}
	
	logrus.WithField("checkpointID", checkpointID).Info("Checkpoint deleted")
	return nil
}

// FindCheckpointByTables finds a checkpoint for specific source and destination tables
func (cm *CheckpointManager) FindCheckpointByTables(sourceTable, destTable string) (*MigrationCheckpoint, error) {
	checkpoints, err := cm.ListCheckpoints()
	if err != nil {
		return nil, err
	}
	
	for _, checkpoint := range checkpoints {
		if checkpoint.SourceTable == sourceTable && checkpoint.DestTable == destTable {
			// Return the most recent matching checkpoint
			return checkpoint, nil
		}
	}
	
	return nil, fmt.Errorf("no checkpoint found for tables %s -> %s", sourceTable, destTable)
}

// GetCheckpointProgress calculates progress percentage
func (checkpoint *MigrationCheckpoint) GetProgress() float64 {
	if checkpoint.TotalRows == 0 {
		return 0
	}
	return float64(checkpoint.ProcessedRows) / float64(checkpoint.TotalRows) * 100
}

// IsResumable checks if a checkpoint can be resumed
func (checkpoint *MigrationCheckpoint) IsResumable() bool {
	return checkpoint.Status == "running" || checkpoint.Status == "paused"
}

// GetRemainingRows returns the number of rows left to process
func (checkpoint *MigrationCheckpoint) GetRemainingRows() int {
	return checkpoint.TotalRows - checkpoint.ProcessedRows
}

// GetEstimatedTimeRemaining estimates remaining time based on current progress
func (checkpoint *MigrationCheckpoint) GetEstimatedTimeRemaining() time.Duration {
	if checkpoint.ProcessedRows == 0 {
		return 0
	}
	
	elapsed := checkpoint.UpdatedAt.Sub(checkpoint.CreatedAt)
	if elapsed == 0 {
		return 0
	}
	
	rowsPerSecond := float64(checkpoint.ProcessedRows) / elapsed.Seconds()
	if rowsPerSecond == 0 {
		return 0
	}
	
	remainingRows := checkpoint.GetRemainingRows()
	remainingSeconds := float64(remainingRows) / rowsPerSecond
	
	return time.Duration(remainingSeconds) * time.Second
}

// UpdateProgress updates the checkpoint with current progress
func (checkpoint *MigrationCheckpoint) UpdateProgress(processedRows int) {
	checkpoint.ProcessedRows = processedRows
	checkpoint.UpdatedAt = time.Now()
	checkpoint.CurrentBatch = processedRows / checkpoint.BatchSize
	
	if processedRows >= checkpoint.TotalRows {
		checkpoint.Status = "completed"
	}
}

// MarkFailed marks the checkpoint as failed with an error message
func (checkpoint *MigrationCheckpoint) MarkFailed(errorMsg string) {
	checkpoint.Status = "failed"
	checkpoint.LastError = errorMsg
	checkpoint.UpdatedAt = time.Now()
}

// MarkPaused marks the checkpoint as paused
func (checkpoint *MigrationCheckpoint) MarkPaused() {
	checkpoint.Status = "paused"
	checkpoint.UpdatedAt = time.Now()
}

// MarkRunning marks the checkpoint as running
func (checkpoint *MigrationCheckpoint) MarkRunning() {
	checkpoint.Status = "running"
	checkpoint.LastError = ""
	checkpoint.UpdatedAt = time.Now()
}
