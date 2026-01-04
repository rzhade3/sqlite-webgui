package models

type Table struct {
	Name       string   `json:"name"`
	RowCount   int      `json:"row_count"`
	ColumnInfo []Column `json:"columns,omitempty"`
}

type Column struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	NotNull      bool    `json:"not_null"`
	DefaultValue *string `json:"default_value"`
	PrimaryKey   bool    `json:"primary_key"`
}

type TableData struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Total   int             `json:"total"`
	Page    int             `json:"page"`
	Limit   int             `json:"limit"`
}

type QueryRequest struct {
	SQL string `json:"sql"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
