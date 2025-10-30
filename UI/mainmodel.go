package UI

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pclubiitk/dbcli/DB"
	"github.com/pclubiitk/dbcli/config"
)

const (
	StepSourceCred = iota
	StepDestCred
	StepSelectSourceTable
	StepSelectSourceColumns
	StepSelectDestTable
	StepSelectDestColumns
	StepMapping
	StepDumpOption
	StepMigrationConfirm
)

// HelpState tracks what help is being shown
type HelpState struct {
	Visible     bool
	CurrentStep int
	Topic       string
}

type Model struct {
	Step int

	// ---- DB CREDENTIAL INPUTS ----
	SourceCred  map[string]string
	DestCred    map[string]string
	Source	    DB.DBInterface  //these are the most imp fields
	Dest        DB.DBInterface  //they are direct connections to databases
	CredInput   textinput.Model
	CredKeys    []string
	CredIndex   int
	IsSource    bool

	// ---- SOURCE DATABASE ----
	SourceTables      []string
	SelectedSourceTbl string
	SourceTableList   list.Model
	SourceColumns     []string
	SelectedSourceCols []string
	SourceColumnList  list.Model
	SourceColSelections map[int]bool  // Track which columns are selected

	// ---- DESTINATION DATABASE ----
	DestTables      []string
	SelectedDestTbl string
	DestTableList   list.Model
	DestColumns     []string
	SelectedDestCols []string
	DestColumnList  list.Model
	DestColSelections map[int]bool    // Track which columns are selected

	// ---- MAPPING ----
	ColumnMapping map[string]string
	CurrentMapIdx int
	MapInput      textinput.Model

	// ---- DUMP OPTION ----
	WantDump    bool
	DumpPath    string
	DumpPathInp textinput.Model

	// ---- MISC ----
	ErrMsg string
	
	// ---- PROGRESS TRACKING ----
	IsProcessing    bool
	ProcessedRows   int
	TotalRows       int
	ProcessingMsg   string

	// ---- CONFIGURATION ----
	Config *config.Config

	// ---- HELP SYSTEM ----
	HelpState *HelpState
}

func (m Model) Init() tea.Cmd {
	// Initialize text input for credentials
	m.CredInput = textinput.New()
	m.CredInput.Focus()
	return nil
}
