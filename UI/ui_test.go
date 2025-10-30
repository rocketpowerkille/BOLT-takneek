package UI

import (
	"testing"
)

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		wantError bool
	}{
		{
			name:      "Valid MySQL vendor",
			key:       "dbVendor",
			value:     "mysql",
			wantError: false,
		},
		{
			name:      "Valid Oracle vendor",
			key:       "dbVendor",
			value:     "oracle",
			wantError: false,
		},
		{
			name:      "Invalid vendor",
			key:       "dbVendor",
			value:     "postgresql",
			wantError: true,
		},
		{
			name:      "Empty vendor",
			key:       "dbVendor",
			value:     "",
			wantError: true,
		},
		{
			name:      "Valid port",
			key:       "port",
			value:     "3306",
			wantError: false,
		},
		{
			name:      "Invalid port - not numeric",
			key:       "port",
			value:     "abc",
			wantError: true,
		},
		{
			name:      "Invalid port - out of range",
			key:       "port",
			value:     "70000",
			wantError: true,
		},
		{
			name:      "Valid localhost",
			key:       "host",
			value:     "localhost",
			wantError: false,
		},
		{
			name:      "Valid IP address",
			key:       "host",
			value:     "192.168.1.1",
			wantError: false,
		},
		{
			name:      "Valid username",
			key:       "user",
			value:     "admin",
			wantError: false,
		},
		{
			name:      "Empty username",
			key:       "user",
			value:     "",
			wantError: true,
		},
		{
			name:      "Valid password",
			key:       "password",
			value:     "secret123",
			wantError: false,
		},
		{
			name:      "Empty password",
			key:       "password",
			value:     "",
			wantError: true,
		},
		{
			name:      "Valid database name",
			key:       "database",
			value:     "mydb",
			wantError: false,
		},
		{
			name:      "Empty database name",
			key:       "database",
			value:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInput(tt.key, tt.value)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error for input '%s'='%s', got nil", tt.key, tt.value)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s'='%s': %v", tt.key, tt.value, err)
				}
			}
		})
	}
}

func TestGoBackStep(t *testing.T) {
	tests := []struct {
		name         string
		currentStep  int
		expectedStep int
	}{
		{
			name:         "Go back from destination credentials",
			currentStep:  StepDestCred,
			expectedStep: StepSourceCred,
		},
		{
			name:         "Go back from source table selection",
			currentStep:  StepSelectSourceTable,
			expectedStep: StepDestCred,
		},
		{
			name:         "Go back from source column selection",
			currentStep:  StepSelectSourceColumns,
			expectedStep: StepSelectSourceTable,
		},
		{
			name:         "Go back from destination table selection",
			currentStep:  StepSelectDestTable,
			expectedStep: StepSelectSourceColumns,
		},
		{
			name:         "Go back from destination column selection",
			currentStep:  StepSelectDestColumns,
			expectedStep: StepSelectDestTable,
		},
		{
			name:         "Go back from mapping",
			currentStep:  StepMapping,
			expectedStep: StepSelectDestColumns,
		},
		{
			name:         "Go back from dump option",
			currentStep:  StepDumpOption,
			expectedStep: StepMapping,
		},
		{
			name:         "Go back from migration confirm",
			currentStep:  StepMigrationConfirm,
			expectedStep: StepDumpOption,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := Model{
				Step: tt.currentStep,
			}
			
			result := goBackStep(model)
			
			if result.Step != tt.expectedStep {
				t.Errorf("Expected step %d, got %d", tt.expectedStep, result.Step)
			}
			
			// Verify error message is cleared
			if result.ErrMsg != "" {
				t.Errorf("Expected empty error message, got '%s'", result.ErrMsg)
			}
		})
	}
}

func TestGetSelectedColumnsFromMap(t *testing.T) {
	columns := []string{"id", "name", "email", "created_at"}
	
	tests := []struct {
		name       string
		selections map[int]bool
		expected   []string
	}{
		{
			name:       "No selections",
			selections: map[int]bool{},
			expected:   []string{},
		},
		{
			name: "Single selection",
			selections: map[int]bool{
				1: true,
			},
			expected: []string{"name"},
		},
		{
			name: "Multiple selections",
			selections: map[int]bool{
				0: true,
				2: true,
				3: true,
			},
			expected: []string{"id", "email", "created_at"},
		},
		{
			name: "All selections",
			selections: map[int]bool{
				0: true,
				1: true,
				2: true,
				3: true,
			},
			expected: []string{"id", "name", "email", "created_at"},
		},
		{
			name: "Mixed selections with false values",
			selections: map[int]bool{
				0: true,
				1: false,
				2: true,
				3: false,
			},
			expected: []string{"id", "email"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSelectedColumnsFromMap(columns, tt.selections)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d selected columns, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected column '%s' at index %d, got '%s'", expected, i, result[i])
				}
			}
		})
	}
}
