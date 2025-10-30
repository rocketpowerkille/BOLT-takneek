package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pclubiitk/dbcli/UI"
	"github.com/pclubiitk/dbcli/config"
	"github.com/pclubiitk/dbcli/logging"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("Warning: Failed to load config, using defaults: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Initialize enhanced logger
	logger, err := logging.NewLogger(cfg)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize logger: %v\n", err)
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		defer logger.Close()
		logrus.Info("DBCLI started with enhanced logging")
	}

	// Initialize the input field
	credInput := textinput.New()
	credInput.Focus()
	credInput.Placeholder = "Enter value..."

	model := UI.Model{
		Step:      UI.StepSourceCred,
		CredKeys:  []string{"dbVendor", "host", "port", "user", "password", "database"},
		IsSource:  true,
		CredInput: credInput,
		CredIndex: 0,
		Config:    cfg, // Pass configuration to UI
	}

	logrus.WithFields(logrus.Fields{
		"version": "1.0.0",
		"config_loaded": cfg != nil,
	}).Info("Starting DBCLI migration tool")

	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		logrus.WithError(err).Error("Failed to start application")
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	logrus.Info("DBCLI session completed")
}
