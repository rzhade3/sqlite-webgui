function app() {
    return {
        loading: false,
        tables: [],
        selectedTable: null,
        tableData: null,
        schema: [],
        currentPage: 1,
        showInsertModal: false,
        showEditModal: false,
        showQueryModal: false,
        newRow: {},
        editingRow: { values: {} },
        customQuery: '',
        queryResult: null,
        darkMode: false,

        async init() {
            this.initDarkMode();
            await this.loadTables();
        },

        initDarkMode() {
            // Check localStorage first, then system preference
            const savedTheme = localStorage.getItem('theme');
            if (savedTheme === 'dark') {
                this.darkMode = true;
                document.documentElement.classList.add('dark');
            } else if (savedTheme === 'light') {
                this.darkMode = false;
                document.documentElement.classList.remove('dark');
            } else {
                // Use system preference
                this.darkMode = window.matchMedia('(prefers-color-scheme: dark)').matches;
                if (this.darkMode) {
                    document.documentElement.classList.add('dark');
                }
            }
        },

        toggleDarkMode() {
            this.darkMode = !this.darkMode;
            if (this.darkMode) {
                document.documentElement.classList.add('dark');
                localStorage.setItem('theme', 'dark');
            } else {
                document.documentElement.classList.remove('dark');
                localStorage.setItem('theme', 'light');
            }
        },

        async loadTables() {
            this.loading = true;
            try {
                const response = await fetch('/api/tables');
                this.tables = await response.json();
            } catch (error) {
                console.error('Failed to load tables:', error);
                alert('Failed to load tables');
            } finally {
                this.loading = false;
            }
        },

        async selectTable(tableName) {
            this.selectedTable = tableName;
            this.currentPage = 1;
            await this.loadSchema();
            await this.loadTableData();
        },

        async loadSchema() {
            try {
                const response = await fetch(`/api/tables/${this.selectedTable}/schema`);
                this.schema = await response.json();
            } catch (error) {
                console.error('Failed to load schema:', error);
            }
        },

        async loadTableData() {
            try {
                const response = await fetch(
                    `/api/tables/${this.selectedTable}/data?page=${this.currentPage}&limit=50`
                );
                this.tableData = await response.json();
            } catch (error) {
                console.error('Failed to load table data:', error);
                alert('Failed to load table data');
            }
        },

        async nextPage() {
            if (this.currentPage * this.tableData.limit < this.tableData.total) {
                this.currentPage++;
                await this.loadTableData();
            }
        },

        async previousPage() {
            if (this.currentPage > 1) {
                this.currentPage--;
                await this.loadTableData();
            }
        },

        editRow(row) {
            this.editingRow.values = {};
            this.schema.forEach((col, idx) => {
                this.editingRow.values[col.name] = row[idx];
            });
            this.showEditModal = true;
        },

        async updateRow() {
            const pkCol = this.schema.find(c => c.primary_key);
            if (!pkCol) {
                alert('No primary key found for this table');
                return;
            }

            const pkValue = this.editingRow.values[pkCol.name];
            const updateData = { ...this.editingRow.values };
            delete updateData[pkCol.name];

            try {
                const response = await fetch(
                    `/api/tables/${this.selectedTable}/rows?pk=${pkCol.name}&pk_value=${pkValue}`,
                    {
                        method: 'PUT',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(updateData)
                    }
                );

                if (response.ok) {
                    this.showEditModal = false;
                    await this.loadTableData();
                    await this.loadTables();
                } else {
                    const error = await response.json();
                    alert('Failed to update row: ' + error.error);
                }
            } catch (error) {
                console.error('Failed to update row:', error);
                alert('Failed to update row');
            }
        },

        async deleteRow(row) {
            const pkCol = this.schema.find(c => c.primary_key);
            if (!pkCol) {
                alert('No primary key found for this table');
                return;
            }

            const pkValue = row[this.schema.findIndex(c => c.primary_key)];
            
            if (!confirm('Are you sure you want to delete this row?')) {
                return;
            }

            try {
                const response = await fetch(
                    `/api/tables/${this.selectedTable}/rows?pk=${pkCol.name}&pk_value=${pkValue}`,
                    { method: 'DELETE' }
                );

                if (response.ok) {
                    await this.loadTableData();
                    await this.loadTables();
                } else {
                    const error = await response.json();
                    alert('Failed to delete row: ' + error.error);
                }
            } catch (error) {
                console.error('Failed to delete row:', error);
                alert('Failed to delete row');
            }
        },

        async insertRow() {
            const rowData = {};
            for (const col of this.schema) {
                const value = this.newRow[col.name];
                if (value !== undefined && value !== '') {
                    rowData[col.name] = value;
                }
            }

            try {
                const response = await fetch(`/api/tables/${this.selectedTable}/rows`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(rowData)
                });

                if (response.ok) {
                    this.showInsertModal = false;
                    this.newRow = {};
                    await this.loadTableData();
                    await this.loadTables();
                } else {
                    const error = await response.json();
                    alert('Failed to insert row: ' + error.error);
                }
            } catch (error) {
                console.error('Failed to insert row:', error);
                alert('Failed to insert row');
            }
        },

        async executeQuery() {
            try {
                const response = await fetch('/api/query', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ sql: this.customQuery })
                });

                if (response.ok) {
                    this.queryResult = await response.json();
                } else {
                    const error = await response.json();
                    alert('Query failed: ' + error.error);
                }
            } catch (error) {
                console.error('Failed to execute query:', error);
                alert('Failed to execute query');
            }
        }
    };
}
