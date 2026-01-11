package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rzhade3/sqlite-webgui/internal/database"
)

func setupTestHandler(t *testing.T, readonly bool) (*APIHandler, string) {
	t.Helper()

	tmpfile, err := os.CreateTemp("", "test*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()

	dbPath := tmpfile.Name()

	setupDB, err := database.New(dbPath, false)
	if err != nil {
		os.Remove(dbPath)
		t.Fatalf("Failed to create setup database: %v", err)
	}

	_, err = setupDB.GetConnection().Exec(`
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

	_, err = setupDB.GetConnection().Exec(`INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com')`)
	if err != nil {
		setupDB.Close()
		os.Remove(dbPath)
		t.Fatalf("Failed to insert test data: %v", err)
	}

	setupDB.Close()

	db, err := database.New(dbPath, readonly)
	if err != nil {
		os.Remove(dbPath)
		t.Fatalf("Failed to open database: %v", err)
	}

	return NewAPIHandler(db), dbPath
}

func TestAPIHandler_InsertRow_ReadOnly(t *testing.T) {
	handler, dbPath := setupTestHandler(t, true)
	defer os.Remove(dbPath)

	body := map[string]interface{}{
		"name":  "Bob",
		"email": "bob@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/tables/users/rows", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	r := chi.NewRouter()
	r.Post("/api/tables/{name}/rows", handler.InsertRow)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	
	if response["error"] != "database is in read-only mode" {
		t.Errorf("Expected 'database is in read-only mode' error, got: %s", response["error"])
	}
}

func TestAPIHandler_InsertRow_Writable(t *testing.T) {
	handler, dbPath := setupTestHandler(t, false)
	defer os.Remove(dbPath)

	body := map[string]interface{}{
		"name":  "Bob",
		"email": "bob@example.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/tables/users/rows", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	r := chi.NewRouter()
	r.Post("/api/tables/{name}/rows", handler.InsertRow)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	
	if response["message"] != "Row inserted successfully" {
		t.Errorf("Expected success message, got: %s", response["message"])
	}
}

func TestAPIHandler_UpdateRow_ReadOnly(t *testing.T) {
	handler, dbPath := setupTestHandler(t, true)
	defer os.Remove(dbPath)

	body := map[string]interface{}{
		"name": "Alice Updated",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/tables/users/rows?pk=id&pk_value=1", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	r := chi.NewRouter()
	r.Put("/api/tables/{name}/rows", handler.UpdateRow)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	
	if response["error"] != "database is in read-only mode" {
		t.Errorf("Expected 'database is in read-only mode' error, got: %s", response["error"])
	}
}

func TestAPIHandler_UpdateRow_Writable(t *testing.T) {
	handler, dbPath := setupTestHandler(t, false)
	defer os.Remove(dbPath)

	body := map[string]interface{}{
		"name": "Alice Updated",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/tables/users/rows?pk=id&pk_value=1", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	r := chi.NewRouter()
	r.Put("/api/tables/{name}/rows", handler.UpdateRow)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	
	if response["message"] != "Row updated successfully" {
		t.Errorf("Expected success message, got: %s", response["message"])
	}
}

func TestAPIHandler_DeleteRow_ReadOnly(t *testing.T) {
	handler, dbPath := setupTestHandler(t, true)
	defer os.Remove(dbPath)

	req := httptest.NewRequest(http.MethodDelete, "/api/tables/users/rows?pk=id&pk_value=1", nil)
	
	w := httptest.NewRecorder()
	
	r := chi.NewRouter()
	r.Delete("/api/tables/{name}/rows", handler.DeleteRow)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	
	if response["error"] != "database is in read-only mode" {
		t.Errorf("Expected 'database is in read-only mode' error, got: %s", response["error"])
	}
}

func TestAPIHandler_DeleteRow_Writable(t *testing.T) {
	handler, dbPath := setupTestHandler(t, false)
	defer os.Remove(dbPath)

	req := httptest.NewRequest(http.MethodDelete, "/api/tables/users/rows?pk=id&pk_value=1", nil)
	
	w := httptest.NewRecorder()
	
	r := chi.NewRouter()
	r.Delete("/api/tables/{name}/rows", handler.DeleteRow)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	
	if response["message"] != "Row deleted successfully" {
		t.Errorf("Expected success message, got: %s", response["message"])
	}
}

func TestAPIHandler_GetTables_ReadOnly(t *testing.T) {
	handler, dbPath := setupTestHandler(t, true)
	defer os.Remove(dbPath)

	req := httptest.NewRequest(http.MethodGet, "/api/tables", nil)
	w := httptest.NewRecorder()
	
	handler.GetTables(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAPIHandler_GetTableData_ReadOnly(t *testing.T) {
	handler, dbPath := setupTestHandler(t, true)
	defer os.Remove(dbPath)

	req := httptest.NewRequest(http.MethodGet, "/api/tables/users/data?page=1&limit=50", nil)
	w := httptest.NewRecorder()
	
	r := chi.NewRouter()
	r.Get("/api/tables/{name}/data", handler.GetTableData)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAPIHandler_GetMode_ReadOnly(t *testing.T) {
	handler, dbPath := setupTestHandler(t, true)
	defer os.Remove(dbPath)

	req := httptest.NewRequest(http.MethodGet, "/api/mode", nil)
	w := httptest.NewRecorder()
	
	handler.GetMode(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]bool
	json.NewDecoder(w.Body).Decode(&response)
	
	if !response["readonly"] {
		t.Error("Expected readonly to be true")
	}
}

func TestAPIHandler_GetMode_Writable(t *testing.T) {
	handler, dbPath := setupTestHandler(t, false)
	defer os.Remove(dbPath)

	req := httptest.NewRequest(http.MethodGet, "/api/mode", nil)
	w := httptest.NewRecorder()
	
	handler.GetMode(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]bool
	json.NewDecoder(w.Body).Decode(&response)
	
	if response["readonly"] {
		t.Error("Expected readonly to be false")
	}
}
