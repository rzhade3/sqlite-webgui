package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rzhade3/sqlite-webgui/internal/database"
	"github.com/rzhade3/sqlite-webgui/internal/models"
)

type APIHandler struct {
	db *database.DB
}

func NewAPIHandler(db *database.DB) *APIHandler {
	return &APIHandler{db: db}
}

func (h *APIHandler) GetTables(w http.ResponseWriter, r *http.Request) {
	tables, err := h.db.GetTables()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tables)
}

func (h *APIHandler) GetMode(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]bool{
		"readonly": h.db.IsReadOnly(),
	})
}

func (h *APIHandler) GetTableSchema(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "name")
	
	schema, err := h.db.GetTableSchema(tableName)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, schema)
}

func (h *APIHandler) GetTableData(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "name")
	
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 1000 {
		limit = 50
	}

	data, err := h.db.GetTableData(tableName, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, data)
}

func (h *APIHandler) InsertRow(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "name")

	var values map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := h.db.InsertRow(tableName, values); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"message": "Row inserted successfully"})
}

func (h *APIHandler) UpdateRow(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "name")
	pkColumn := r.URL.Query().Get("pk")
	pkValue := r.URL.Query().Get("pk_value")

	if pkColumn == "" || pkValue == "" {
		respondError(w, http.StatusBadRequest, "Missing pk or pk_value query parameters")
		return
	}

	var values map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := h.db.UpdateRow(tableName, pkColumn, pkValue, values); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Row updated successfully"})
}

func (h *APIHandler) DeleteRow(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "name")
	pkColumn := r.URL.Query().Get("pk")
	pkValue := r.URL.Query().Get("pk_value")

	if pkColumn == "" || pkValue == "" {
		respondError(w, http.StatusBadRequest, "Missing pk or pk_value query parameters")
		return
	}

	if err := h.db.DeleteRow(tableName, pkColumn, pkValue); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Row deleted successfully"})
}

func (h *APIHandler) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	var req models.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	data, err := h.db.ExecuteQuery(req.SQL)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, data)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, models.ErrorResponse{Error: message})
}
