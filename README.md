# SQLite Web GUI

A fast, modern web-based GUI for viewing and editing SQLite databases. Built with Go, Chi router, and Alpine.js.

## Features

**View all tables** - Browse all tables in your SQLite database  
**View data** - See all rows with pagination (50 rows per page)  
**CRUD Operations** - Create, Read, Update, and Delete rows  
**Schema inspection** - View column types, constraints, and primary keys  
**Custom SQL execution** - Execute any SQL query (SELECT, UPDATE, INSERT, DELETE)  
**Read-only mode** - Safe default mode that prevents accidental data modification  
**Writable mode** - Optional mode with full write access for data editing

## Quick Start

### Download and Run

1. Download the binary for your platform from releases
2. Run the application:

```bash
# Read-only mode (default - safe for browsing)
./sqlite-webgui mydatabase.db

# Writable mode (enables data modification)
./sqlite-webgui --writable mydatabase.db
```

3. Open your browser to `http://localhost:8080`

### Custom Port

```bash
./sqlite-webgui --port 3000 mydatabase.db
./sqlite-webgui --port 3000 --writable mydatabase.db
```

## Building from Source

### Prerequisites

- Go 1.21 or later

### Build

```bash
# Clone the repository
git clone https://github.com/rzhade3/sqlite-webgui.git
cd sqlite-webgui

# Install dependencies
go mod download

# Build
go build -o sqlite-webgui .

# Run
./sqlite-webgui sample.db
```

### Cross-Compile for Other Platforms

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o sqlite-webgui.exe .

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o sqlite-webgui-mac-intel .

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o sqlite-webgui-mac-arm .

# Linux
GOOS=linux GOARCH=amd64 go build -o sqlite-webgui-linux .
```

## Usage

### Basic Usage

```bash
# Read-only mode (default - safe for data browsing)
./sqlite-webgui mydata.db

# Writable mode (enables data modification)
./sqlite-webgui --writable mydata.db

# Custom port
./sqlite-webgui --port 3000 mydata.db
./sqlite-webgui --port 3000 --writable mydata.db
```

### Modes

**Read-only mode (default):**
- Safe for browsing and analyzing data
- Prevents accidental data modification
- SQL queries limited to SELECT statements
- No UI buttons for Add, Edit, or Delete

**Writable mode (--writable flag):**
- Full data editing capabilities
- SQL queries can execute UPDATE, INSERT, DELETE
- UI shows all CRUD operation buttons
- Use with caution on production databases

### Web Interface

Once the server is running, open your browser to the URL shown in the terminal (default: `http://localhost:8080`).

**Sidebar:**
- View all tables in the database
- See row counts for each table
- Click to select a table
- Mode indicator badge (READ-ONLY or READ-WRITE)

**Main View:**
- Browse table data with pagination
- Click "Add Row" to insert new records (writable mode only)
- Click "Edit" to modify existing rows (writable mode only)
- Click "Delete" to remove rows (writable mode only)
- Use "Execute SQL" to run custom queries (SELECT in readonly, any SQL in writable)

### API Endpoints

The application provides a REST API:

```
GET    /api/mode                        - Get current mode (readonly status)
GET    /api/tables                      - List all tables
GET    /api/tables/:name/schema         - Get table schema
GET    /api/tables/:name/data           - Get table data (paginated)
POST   /api/query                       - Execute SQL query
POST   /api/tables/:name/rows           - Insert a new row (writable mode only)
PUT    /api/tables/:name/rows           - Update a row (writable mode only)
DELETE /api/tables/:name/rows           - Delete a row (writable mode only)
```

Example:

```bash
# Get all tables
curl http://localhost:8080/api/tables

# Get data from users table
curl "http://localhost:8080/api/tables/users/data?page=1&limit=50"

# Insert a new row
curl -X POST http://localhost:8080/api/tables/users/rows \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com", "age": 30}'
```

## Security Notes

⚠️ **Important:** This tool is designed for **local development and testing**. 

**Read-only mode (default):**
- Database opened with SQLite's read-only flag
- Safe for browsing production databases
- No risk of accidental data modification
- SQL execution limited to SELECT by SQLite itself

**Writable mode (--writable flag):**
- Full database write access
- All SQL operations allowed (UPDATE, INSERT, DELETE)
- No authentication or authorization
- No rate limiting
- **Do not expose to the internet without proper security measures**
- **Use with caution on production databases**

## Roadmap

Future enhancements:
- [ ] Import/Export CSV
- [ ] Multi-row selection and bulk operations
- [ ] Query history
- [ ] Table creation/modification
- [ ] Index management
- [ ] Full-text search
