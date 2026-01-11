package database

import (
	"os"
	"strings"
	"testing"
)

func setupTestDB(t *testing.T, readonly bool) (*DB, string) {
	t.Helper()
	
	tmpfile, err := os.CreateTemp("", "test*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()
	
	dbPath := tmpfile.Name()
	
	// Create a writable connection to set up the schema
	setupDB, err := New(dbPath, false)
	if err != nil {
		os.Remove(dbPath)
		t.Fatalf("Failed to create setup database: %v", err)
	}
	
	_, err = setupDB.conn.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT
		)
	`)
	if err != nil {
		setupDB.Close()
		os.Remove(dbPath)
		t.Fatalf("Failed to create schema: %v", err)
	}
	
	_, err = setupDB.conn.Exec(`INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com'), ('Bob', 'bob@example.com')`)
	if err != nil {
		setupDB.Close()
		os.Remove(dbPath)
		t.Fatalf("Failed to insert test data: %v", err)
	}
	
	setupDB.Close()
	
	// Open with the requested mode
	db, err := New(dbPath, readonly)
	if err != nil {
		os.Remove(dbPath)
		t.Fatalf("Failed to open database: %v", err)
	}
	
	return db, dbPath
}

func TestNew_ReadOnly(t *testing.T) {
	db, dbPath := setupTestDB(t, true)
	defer db.Close()
	defer os.Remove(dbPath)
	
	if !db.IsReadOnly() {
		t.Error("Expected database to be in read-only mode")
	}
}

func TestNew_Writable(t *testing.T) {
	db, dbPath := setupTestDB(t, false)
	defer db.Close()
	defer os.Remove(dbPath)
	
	if db.IsReadOnly() {
		t.Error("Expected database to be in writable mode")
	}
}

func TestInsertRow_ReadOnly(t *testing.T) {
	db, dbPath := setupTestDB(t, true)
	defer db.Close()
	defer os.Remove(dbPath)
	
	values := map[string]interface{}{
		"name":  "Charlie",
		"email": "charlie@example.com",
	}
	
	err := db.InsertRow("users", values)
	if err == nil {
		t.Error("Expected error when inserting in read-only mode")
	}
	
	if err.Error() != "database is in read-only mode" {
		t.Errorf("Expected 'database is in read-only mode' error, got: %v", err)
	}
}

func TestInsertRow_Writable(t *testing.T) {
	db, dbPath := setupTestDB(t, false)
	defer db.Close()
	defer os.Remove(dbPath)
	
	values := map[string]interface{}{
		"name":  "Charlie",
		"email": "charlie@example.com",
	}
	
	err := db.InsertRow("users", values)
	if err != nil {
		t.Errorf("Expected no error when inserting in writable mode, got: %v", err)
	}
	
	// Verify the row was inserted
	data, err := db.GetTableData("users", 1, 50)
	if err != nil {
		t.Fatalf("Failed to get table data: %v", err)
	}
	
	if data.Total != 3 {
		t.Errorf("Expected 3 rows, got %d", data.Total)
	}
}

func TestUpdateRow_ReadOnly(t *testing.T) {
	db, dbPath := setupTestDB(t, true)
	defer db.Close()
	defer os.Remove(dbPath)
	
	values := map[string]interface{}{
		"name": "Alice Updated",
	}
	
	err := db.UpdateRow("users", "id", 1, values)
	if err == nil {
		t.Error("Expected error when updating in read-only mode")
	}
	
	if err.Error() != "database is in read-only mode" {
		t.Errorf("Expected 'database is in read-only mode' error, got: %v", err)
	}
}

func TestUpdateRow_Writable(t *testing.T) {
	db, dbPath := setupTestDB(t, false)
	defer db.Close()
	defer os.Remove(dbPath)
	
	values := map[string]interface{}{
		"name": "Alice Updated",
	}
	
	err := db.UpdateRow("users", "id", 1, values)
	if err != nil {
		t.Errorf("Expected no error when updating in writable mode, got: %v", err)
	}
	
	// Verify the row was updated
	data, err := db.GetTableData("users", 1, 50)
	if err != nil {
		t.Fatalf("Failed to get table data: %v", err)
	}
	
	if len(data.Rows) > 0 {
		name := data.Rows[0][1].(string)
		if name != "Alice Updated" {
			t.Errorf("Expected name 'Alice Updated', got '%s'", name)
		}
	}
}

func TestDeleteRow_ReadOnly(t *testing.T) {
	db, dbPath := setupTestDB(t, true)
	defer db.Close()
	defer os.Remove(dbPath)
	
	err := db.DeleteRow("users", "id", 1)
	if err == nil {
		t.Error("Expected error when deleting in read-only mode")
	}
	
	if err.Error() != "database is in read-only mode" {
		t.Errorf("Expected 'database is in read-only mode' error, got: %v", err)
	}
}

func TestDeleteRow_Writable(t *testing.T) {
	db, dbPath := setupTestDB(t, false)
	defer db.Close()
	defer os.Remove(dbPath)
	
	err := db.DeleteRow("users", "id", 1)
	if err != nil {
		t.Errorf("Expected no error when deleting in writable mode, got: %v", err)
	}
	
	// Verify the row was deleted
	data, err := db.GetTableData("users", 1, 50)
	if err != nil {
		t.Fatalf("Failed to get table data: %v", err)
	}
	
	if data.Total != 1 {
		t.Errorf("Expected 1 row after deletion, got %d", data.Total)
	}
}

func TestGetTables_ReadOnly(t *testing.T) {
	db, dbPath := setupTestDB(t, true)
	defer db.Close()
	defer os.Remove(dbPath)
	
	tables, err := db.GetTables()
	if err != nil {
		t.Errorf("Expected no error when getting tables in read-only mode, got: %v", err)
	}
	
	if len(tables) != 1 || tables[0].Name != "users" {
		t.Errorf("Expected 1 table named 'users', got: %v", tables)
	}
}

func TestGetTableData_ReadOnly(t *testing.T) {
	db, dbPath := setupTestDB(t, true)
	defer db.Close()
	defer os.Remove(dbPath)
	
	data, err := db.GetTableData("users", 1, 50)
	if err != nil {
		t.Errorf("Expected no error when reading data in read-only mode, got: %v", err)
	}
	
	if data.Total != 2 {
		t.Errorf("Expected 2 rows, got %d", data.Total)
	}
}

func TestExecuteQuery_ReadOnly(t *testing.T) {
	db, dbPath := setupTestDB(t, true)
	defer db.Close()
	defer os.Remove(dbPath)
	
	data, err := db.ExecuteQuery("SELECT * FROM users WHERE name = 'Alice'")
	if err != nil {
		t.Errorf("Expected no error when executing SELECT in read-only mode, got: %v", err)
	}
	
	if len(data.Rows) != 1 {
		t.Errorf("Expected 1 row in query result, got %d", len(data.Rows))
	}
}

func TestExecuteQuery_WriteInReadOnly(t *testing.T) {
	db, dbPath := setupTestDB(t, true)
	defer db.Close()
	defer os.Remove(dbPath)
	
	// Attempt UPDATE in readonly mode - should fail at SQLite level
	_, err := db.ExecuteQuery("UPDATE users SET name = 'Hacked' WHERE id = 1")
	if err == nil {
		t.Error("Expected error when executing UPDATE in read-only mode")
	}
	
	if !strings.Contains(err.Error(), "readonly") {
		t.Errorf("Expected 'readonly' in error message, got: %v", err)
	}
}

func TestExecuteQuery_WriteInWritable(t *testing.T) {
	db, dbPath := setupTestDB(t, false)
	defer db.Close()
	defer os.Remove(dbPath)
	
	// Execute UPDATE in writable mode - should succeed
	_, err := db.ExecuteQuery("UPDATE users SET name = 'Alice Updated' WHERE id = 1")
	if err != nil {
		t.Errorf("Expected no error when executing UPDATE in writable mode, got: %v", err)
	}
	
	// Verify the update worked
	data, err := db.ExecuteQuery("SELECT name FROM users WHERE id = 1")
	if err != nil {
		t.Fatalf("Failed to query after update: %v", err)
	}
	
	if len(data.Rows) > 0 {
		name := data.Rows[0][0].(string)
		if name != "Alice Updated" {
			t.Errorf("Expected name 'Alice Updated', got '%s'", name)
		}
	}
}
