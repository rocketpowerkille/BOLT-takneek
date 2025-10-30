package main

import (
	"os"
	"testing"

	"github.com/pclubiitk/dbcli/config"
	"github.com/pclubiitk/dbcli/logging"
)

func TestMainFunctionality(t *testing.T) {
	// Test configuration loading
	t.Run("TestConfigurationLoading", func(t *testing.T) {
		cfg, err := config.LoadConfig("")
		if err != nil {
			// Should fall back to default config
			cfg = config.DefaultConfig()
		}
		
		if cfg == nil {
			t.Fatal("Configuration should not be nil")
		}
		
		if cfg.Migration.DefaultBatchSize <= 0 {
			t.Error("Batch size should be positive")
		}
		
		if cfg.Database.ConnectionTimeout < 0 {
			t.Error("Connection timeout should not be negative")
		}
	})
	
	// Test enhanced logger initialization
	t.Run("TestLoggerInitialization", func(t *testing.T) {
		cfg := config.DefaultConfig()
		
		// Create a temporary log directory for testing
		tmpDir := "test_logs"
		cfg.Logging.LogFilePath = tmpDir + "/test.log"
		os.MkdirAll(tmpDir, 0755)
		defer os.RemoveAll(tmpDir)
		
		logger, err := logging.NewLogger(cfg)
		if err != nil {
			t.Errorf("Logger initialization failed: %v", err)
			return
		}
		defer logger.Close()
		
		// Test that logger is functional
		logger.Info("Test log message")
	})
	
	// Test configuration with various settings
	t.Run("TestConfigurationDefaults", func(t *testing.T) {
		cfg := config.DefaultConfig()
		
		// Verify default values
		if cfg.Migration.DefaultBatchSize != 1000 {
			t.Errorf("Expected batch size 1000, got %d", cfg.Migration.DefaultBatchSize)
		}
		
		if cfg.Database.ConnectionTimeout != 30 {
			t.Errorf("Expected connection timeout 30s, got %d", cfg.Database.ConnectionTimeout)
		}
		
		if !cfg.Migration.EnableProgress {
			t.Error("Progress should be enabled by default")
		}
		
		if !cfg.Migration.EnableTransactions {
			t.Error("Transactions should be enabled by default")
		}
	})
}

func TestProductionReadiness(t *testing.T) {
	t.Run("TestAllComponentsIntegrated", func(t *testing.T) {
		// Verify all critical components can be initialized
		cfg := config.DefaultConfig()
		
		// Test logger
		tmpDir := "test_logs_prod"
		cfg.Logging.LogFilePath = tmpDir + "/prod_test.log"
		os.MkdirAll(tmpDir, 0755)
		defer os.RemoveAll(tmpDir)
		
		logger, err := logging.NewLogger(cfg)
		if err != nil {
			t.Errorf("Production logger setup failed: %v", err)
		} else {
			logger.Close()
		}
		
		// Test that configuration can be saved and loaded
		configPath := tmpDir + "/config.yaml"
		err = cfg.SaveConfig(configPath)
		if err != nil {
			t.Errorf("Config save failed: %v", err)
		}
		
		loadedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			t.Errorf("Config load failed: %v", err)
		}
		
		if loadedCfg.Migration.DefaultBatchSize != cfg.Migration.DefaultBatchSize {
			t.Error("Configuration persistence failed")
		}
	})
}
