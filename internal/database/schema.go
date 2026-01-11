package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/rzhade3/sqlite-webgui/internal/models"
)

func (db *DB) GetTables() ([]models.Table, error) {
	query := `
		SELECT name 
		FROM sqlite_master 
		WHERE type='table' 
		AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []models.Table
	for rows.Next() {
		var table models.Table
		if err := rows.Scan(&table.Name); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}

		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", table.Name)
		if err := db.conn.QueryRow(countQuery).Scan(&table.RowCount); err != nil {
			table.RowCount = 0
		}

		tables = append(tables, table)
	}

	return tables, rows.Err()
}

func (db *DB) GetTableSchema(tableName string) ([]models.Column, error) {
	query := fmt.Sprintf("PRAGMA table_info(`%s`)", tableName)
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table schema: %w", err)
	}
	defer rows.Close()

	var columns []models.Column
	for rows.Next() {
		var (
			cid          int
			name         string
			colType      string
			notNull      int
			defaultValue sql.NullString
			pk           int
		)

		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultValue, &pk); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		col := models.Column{
			Name:       name,
			Type:       colType,
			NotNull:    notNull == 1,
			PrimaryKey: pk > 0,
		}

		if defaultValue.Valid {
			col.DefaultValue = &defaultValue.String
		}

		columns = append(columns, col)
	}

	return columns, rows.Err()
}

func (db *DB) GetTableData(tableName string, page, limit int) (*models.TableData, error) {
	offset := (page - 1) * limit

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	if err := db.conn.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count rows: %w", err)
	}

	dataQuery := fmt.Sprintf("SELECT * FROM `%s` LIMIT ? OFFSET ?", tableName)
	rows, err := db.conn.Query(dataQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query table data: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var data [][]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		for i, v := range values {
			if b, ok := v.([]byte); ok {
				values[i] = string(b)
			}
		}

		data = append(data, values)
	}

	return &models.TableData{
		Columns: columns,
		Rows:    data,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, rows.Err()
}

func (db *DB) InsertRow(tableName string, values map[string]interface{}) error {
	if db.readonly {
		return fmt.Errorf("database is in read-only mode")
	}

	var columns []string
	var placeholders []string
	var args []interface{}

	for col, val := range values {
		columns = append(columns, fmt.Sprintf("`%s`", col))
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}

	query := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := db.conn.Exec(query, args...)
	return err
}

func (db *DB) UpdateRow(tableName string, pkColumn string, pkValue interface{}, values map[string]interface{}) error {
	if db.readonly {
		return fmt.Errorf("database is in read-only mode")
	}

	var setClauses []string
	var args []interface{}

	for col, val := range values {
		setClauses = append(setClauses, fmt.Sprintf("`%s` = ?", col))
		args = append(args, val)
	}

	args = append(args, pkValue)

	query := fmt.Sprintf(
		"UPDATE `%s` SET %s WHERE `%s` = ?",
		tableName,
		strings.Join(setClauses, ", "),
		pkColumn,
	)

	_, err := db.conn.Exec(query, args...)
	return err
}

func (db *DB) DeleteRow(tableName string, pkColumn string, pkValue interface{}) error {
	if db.readonly {
		return fmt.Errorf("database is in read-only mode")
	}

	query := fmt.Sprintf("DELETE FROM `%s` WHERE `%s` = ?", tableName, pkColumn)
	_, err := db.conn.Exec(query, pkValue)
	return err
}

func (db *DB) ExecuteQuery(query string) (*models.TableData, error) {
	query = strings.TrimSpace(query)
	
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var data [][]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		for i, v := range values {
			if b, ok := v.([]byte); ok {
				values[i] = string(b)
			}
		}

		data = append(data, values)
	}

	return &models.TableData{
		Columns: columns,
		Rows:    data,
		Total:   len(data),
		Page:    1,
		Limit:   len(data),
	}, rows.Err()
}
