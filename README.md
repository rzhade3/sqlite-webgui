# SQLite Web GUI

A fast, modern web-based GUI for viewing and editing SQLite databases. Built with Go, Chi router, and Alpine.js.

## Features

**View all tables** - Browse all tables in your SQLite database  
**View data** - See all rows with pagination (50 rows per page)  
**CRUD Operations** - Create, Read, Update, and Delete rows  
**Schema inspection** - View column types, constraints, and primary keys  
**Custom SQL queries** - Execute SELECT queries directly  

## Quick Start

### Download and Run

1. Download the binary for your platform from releases
2. Run the application:

```bash
./sqlite-webgui mydatabase.db
```

3. Open your browser to `http://localhost:8080`

### Custom Port

```bash
./sqlite-webgui --port 3000 mydatabase.db
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
# Start the web server with a database file
./sqlite-webgui mydata.db

# Specify a custom port
./sqlite-webgui --port 3000 mydata.db
```

### Web Interface

Once the server is running, open your browser to the URL shown in the terminal (default: `http://localhost:8080`).

**Sidebar:**
- View all tables in the database
- See row counts for each table
- Click to select a table

**Main View:**
- Browse table data with pagination
- Click "Add Row" to insert new records
- Click "Edit" to modify existing rows
- Click "Delete" to remove rows
- Use "Execute SQL Query" to run custom SELECT queries

### API Endpoints

The application provides a REST API:

```
GET    /api/tables                      - List all tables
GET    /api/tables/:name/schema         - Get table schema
GET    /api/tables/:name/data           - Get table data (paginated)
POST   /api/tables/:name/rows           - Insert a new row
PUT    /api/tables/:name/rows           - Update a row
DELETE /api/tables/:name/rows           - Delete a row
POST   /api/query                       - Execute a SELECT query
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

- All database operations are allowed (INSERT, UPDATE, DELETE)
- No authentication or authorization
- No rate limiting
- Only SELECT queries are allowed via the custom query interface
- **Do not expose this to the internet without proper security measures**

## Roadmap

Future enhancements:
- [ ] Import/Export CSV
- [ ] Multi-row selection and bulk operations
- [ ] Query history
- [ ] Table creation/modification
- [ ] Index management
- [ ] Full-text search
