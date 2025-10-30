# DBCLI - Database Migration Tool

🚀 **A powerful, user-friendly database migration tool with enterprise-grade features**

## Quick Start

### 1. Run the Application
```bash
# Windows
dbcli.exe

# Linux/Mac (if built from source)
./dbcli
```

### 2. Follow the Interactive Setup
1. **Source Database**: Enter credentials for your source database
2. **Destination Database**: Enter credentials for your destination database  
3. **Select Tables**: Choose which tables to migrate
4. **Select Columns**: Pick specific columns to migrate
5. **Map Columns**: Define how source columns map to destination columns
6. **Configure Options**: Set dump options and migration settings
7. **Execute**: Review and start the migration

## Supported Databases
- ✅ **MySQL** (5.7+, 8.0+)
- ✅ **Oracle** (11g+)

## Key Features

### Core Migration
- 🔄 **Interactive TUI** - Step-by-step guided process
- 🎯 **Column Mapping** - Flexible source-to-destination mapping
- 📊 **Progress Tracking** - Real-time migration progress
- 💾 **SQL Dumps** - Generate backup scripts

### Enterprise Features
- ⚙️ **Configuration Management** - YAML/JSON configuration files
- 🔌 **Connection Pooling** - Optimized database connections
- ✅ **Data Validation** - Pre-migration integrity checks
- 📚 **Help System** - Context-sensitive guidance
- 🔄 **Checkpoint/Resume** - Resume interrupted migrations
- 📝 **Enhanced Logging** - Structured logging with rotation
- 🛡️ **Error Recovery** - Automatic retries and rollback
- 📈 **Performance Monitoring** - Migration metrics and timing
- 📋 **Migration Templates** - Reusable configuration patterns
- 🔍 **Dry-Run Mode** - Preview migrations before execution

## Installation

1. Clone the repository:
```bash
git clone https://github.com/pclubiitk/dbcli.git
cd dbcli
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o dbcli main.go
```

## Usage

Run the application:
```bash
./dbcli
```

### Migration Process

1. **Source Database Credentials**: Enter connection details for your source database
   - Database Vendor: `mysql` or `oracle`
   - Host: Database server hostname or IP
   - Port: Database port (3306 for MySQL, 1521 for Oracle)
   - Username: Database username
   - Password: Database password (masked input)
   - Database: Database name (MySQL) or Service name (Oracle)

2. **Destination Database Credentials**: Same process for destination database

3. **Table Selection**: Choose source and destination tables from interactive lists

4. **Column Selection**: Select specific columns to migrate
   - Use ↑↓ to navigate
   - Space to toggle individual columns
   - 'a' to select all columns
   - 'n' to deselect all columns

5. **Column Mapping**: Map source columns to destination columns interactively

6. **Dump Configuration**: Optionally generate SQL dump files

7. **Migration Confirmation**: Review and confirm migration details

8. **Migration Execution**: Data is migrated with progress feedback

## Supported Databases

- **MySQL** (using GORM)
- **Oracle** (using godror driver)

## Controls

- **↑↓ Arrow Keys**: Navigate lists and options
- **Enter**: Confirm selection and proceed
- **Space**: Toggle column selections
- **Esc**: Go back to previous field (in credential input)
- **a**: Select all columns
- **n**: Deselect all columns
- **Ctrl+C/q**: Quit application

## Configuration Examples

### MySQL Connection
- Vendor: `mysql`
- Host: `localhost`
- Port: `3306`
- Username: `root`
- Password: `your_password`
- Database: `mydb`

### Oracle Connection
- Vendor: `oracle`
- Host: `localhost`
- Port: `1521`
- Username: `hr`
- Password: `your_password`
- Database: `XE` (service name)

## Error Handling

The application provides comprehensive error handling for:
- Invalid database credentials
- Connection failures
- Missing required fields
- Invalid hostnames/ports
- Column mapping mismatches

## Development

### Dependencies

- Go 1.24.4+
- github.com/charmbracelet/bubbletea (TUI framework)
- github.com/charmbracelet/bubbles (TUI components)
- gorm.io/gorm (MySQL ORM)
- github.com/godror/godror (Oracle driver)
- github.com/sirupsen/logrus (Logging)

### Building from Source

```bash
go mod download
go build -o dbcli main.go
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and questions, please open an issue on the GitHub repository.