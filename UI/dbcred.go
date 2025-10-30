package UI

import (
	"fmt"
	"strings"
	"strconv"
	"net"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pclubiitk/dbcli/DB"
)

func UpdateDBCred(m Model, msg tea.Msg) Model {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Validate input before saving
			currentKey := m.CredKeys[m.CredIndex]
			inputValue := m.CredInput.Value()
			
			if err := validateInput(currentKey, inputValue); err != nil {
				m.ErrMsg = err.Error()
				return m
			}
			
			// Save input
			if m.IsSource {
				if m.SourceCred == nil {
					m.SourceCred = make(map[string]string)
				}
				m.SourceCred[currentKey] = inputValue
			} else {
				if m.DestCred == nil {
					m.DestCred = make(map[string]string)
				}
				m.DestCred[currentKey] = inputValue
			}

			// Reset input and error
			m.CredInput.SetValue("")
			m.ErrMsg = ""
			m.CredIndex++
			
			if m.CredIndex >= len(m.CredKeys) {
				m.CredIndex = 0
				
				// Test connection before proceeding
				if m.IsSource {
					if err := testAndConnectSource(m); err != nil {
						m.ErrMsg = fmt.Sprintf("Source connection failed: %v", err)
						m.CredIndex = 0 // Reset to start of credentials
						return m
					}
				} else {
					if err := testAndConnectDest(m); err != nil {
						m.ErrMsg = fmt.Sprintf("Destination connection failed: %v", err)
						m.CredIndex = 0 // Reset to start of credentials
						return m
					}
				}
				
				m.Step++
				m.IsSource = !m.IsSource
			}
		case "backspace":
			if len(m.CredInput.Value()) > 0 {
				m.CredInput.SetValue(m.CredInput.Value()[:len(m.CredInput.Value())-1])
			}
		case "ctrl+c", "esc":
			// Go back to previous field
			if m.CredIndex > 0 {
				m.CredIndex--
				// Restore previous value
				var prevValue string
				if m.IsSource && m.SourceCred != nil {
					prevValue = m.SourceCred[m.CredKeys[m.CredIndex]]
				} else if !m.IsSource && m.DestCred != nil {
					prevValue = m.DestCred[m.CredKeys[m.CredIndex]]
				}
				m.CredInput.SetValue(prevValue)
				m.ErrMsg = ""
			}
		default:
			// For password fields, don't echo the characters
			if m.CredKeys[m.CredIndex] == "password" {
				m.CredInput.SetValue(m.CredInput.Value() + msg.String())
			} else {
				m.CredInput.SetValue(m.CredInput.Value() + msg.String())
			}
		}
	}
	return m
}

func validateInput(key, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", key)
	}
	
	switch key {
	case "dbVendor":
		vendor := strings.ToLower(value)
		if vendor != "mysql" && vendor != "oracle" {
			return fmt.Errorf("unsupported database vendor. Use 'mysql' or 'oracle'")
		}
	case "host":
		// Validate hostname or IP
		if net.ParseIP(value) == nil {
			// If not IP, check if it's a valid hostname
			if _, err := net.LookupHost(value); err != nil && value != "localhost" {
				return fmt.Errorf("invalid hostname or IP address")
			}
		}
	case "port":
		// Validate port number
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("port must be a number")
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}
	case "user":
		if len(value) < 1 {
			return fmt.Errorf("username cannot be empty")
		}
	case "password":
		if len(value) < 1 {
			return fmt.Errorf("password cannot be empty")
		}
	case "database":
		if len(value) < 1 {
			return fmt.Errorf("database/service name cannot be empty")
		}
	}
	
	return nil
}

func testAndConnectSource(m Model) error {
	loweredVendor := strings.ToLower(m.SourceCred["dbVendor"])
	switch loweredVendor {
	case "oracle":
		if err := DB.TestOracleConnection(m.SourceCred["host"], m.SourceCred["port"], m.SourceCred["database"], m.SourceCred["user"], m.SourceCred["password"]); err != nil {
			return err
		}
		DB.ConnectOracle(m.SourceCred["host"], m.SourceCred["port"], m.SourceCred["database"], m.SourceCred["user"], m.SourceCred["password"])
		m.Source = &DB.SQLWrapper{DB: DB.OracleDB}
	case "mysql":
		if err := DB.TestMySQLConnection(m.SourceCred["host"], m.SourceCred["port"], m.SourceCred["password"], m.SourceCred["database"], m.SourceCred["user"]); err != nil {
			return err
		}
		DB.ConnectMySQL(m.SourceCred["host"], m.SourceCred["port"], m.SourceCred["password"], m.SourceCred["database"], m.SourceCred["user"])
		m.Source = &DB.GormWrapper{DB: DB.MySQLDB}
	default:
		return fmt.Errorf("unsupported database vendor: %s", loweredVendor)
	}
	return nil
}

func testAndConnectDest(m Model) error {
	loweredVendor := strings.ToLower(m.DestCred["dbVendor"])
	switch loweredVendor {
	case "oracle":
		if err := DB.TestOracleConnection(m.DestCred["host"], m.DestCred["port"], m.DestCred["database"], m.DestCred["user"], m.DestCred["password"]); err != nil {
			return err
		}
		DB.ConnectOracle(m.DestCred["host"], m.DestCred["port"], m.DestCred["database"], m.DestCred["user"], m.DestCred["password"])
		m.Dest = &DB.SQLWrapper{DB: DB.OracleDB}
	case "mysql":
		if err := DB.TestMySQLConnection(m.DestCred["host"], m.DestCred["port"], m.DestCred["password"], m.DestCred["database"], m.DestCred["user"]); err != nil {
			return err
		}
		DB.ConnectMySQL(m.DestCred["host"], m.DestCred["port"], m.DestCred["password"], m.DestCred["database"], m.DestCred["user"])
		m.Dest = &DB.GormWrapper{DB: DB.MySQLDB}
	default:
		return fmt.Errorf("unsupported database vendor: %s", loweredVendor)
	}
	return nil
}

func ViewDBCred(m Model) string {
	var sb strings.Builder
	dbType := "Source"
	if !m.IsSource {
		dbType = "Destination"
	}

	sb.WriteString(fmt.Sprintf("Step: Enter %s DB Credentials\n", dbType))
	sb.WriteString("=====================================\n\n")
	
	// Show progress
	sb.WriteString(fmt.Sprintf("Progress: %d/%d\n\n", m.CredIndex+1, len(m.CredKeys)))
	
	if m.CredIndex < len(m.CredKeys) {
		key := m.CredKeys[m.CredIndex]
		
		// Show helpful hints for each field
		switch key {
		case "dbVendor":
			sb.WriteString("Database Vendor (mysql/oracle): ")
		case "host":
			sb.WriteString("Host (localhost or IP address): ")
		case "port":
			sb.WriteString("Port (3306 for MySQL, 1521 for Oracle): ")
		case "user":
			sb.WriteString("Username: ")
		case "password":
			sb.WriteString("Password: ")
		case "database":
			vendor := ""
			if m.IsSource && m.SourceCred != nil {
				vendor = m.SourceCred["dbVendor"]
			} else if !m.IsSource && m.DestCred != nil {
				vendor = m.DestCred["dbVendor"]
			}
			if strings.ToLower(vendor) == "oracle" {
				sb.WriteString("Service Name (e.g., XE, ORCL): ")
			} else {
				sb.WriteString("Database Name: ")
			}
		}
		
		// Mask password input
		if key == "password" {
			maskedValue := strings.Repeat("*", len(m.CredInput.Value()))
			sb.WriteString(maskedValue)
		} else {
			sb.WriteString(m.CredInput.Value())
		}
		
		sb.WriteString("\n\nType your input and press Enter to continue.")
		sb.WriteString("\nPress Esc to go back to previous field.")
	} else {
		sb.WriteString("All credentials entered. Testing connection...")
	}

	if m.ErrMsg != "" {
		sb.WriteString(fmt.Sprintf("\n\n❌ Error: %s", m.ErrMsg))
	}

	sb.WriteString("\n\nNavigation:")
	if m.Step > StepSourceCred {
		sb.WriteString("\n  Ctrl+B - Go back to previous step")
	}
	sb.WriteString("\n  Ctrl+C - Quit application")
	return sb.String()
}
