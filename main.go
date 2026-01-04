package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rzhade3/sqlite-webgui/internal/database"
	"github.com/rzhade3/sqlite-webgui/internal/handlers"
)

//go:embed internal/handlers/web/*
var webFS embed.FS

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--port PORT] <database.db>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s mydata.db\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --port 3000 mydata.db\n", os.Args[0])
		os.Exit(1)
	}

	dbPath := args[0]

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatalf("Database file does not exist: %s", dbPath)
	}

	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	apiHandler := handlers.NewAPIHandler(db)

	r.Route("/api", func(r chi.Router) {
		r.Get("/tables", apiHandler.GetTables)
		r.Get("/tables/{name}/schema", apiHandler.GetTableSchema)
		r.Get("/tables/{name}/data", apiHandler.GetTableData)
		r.Post("/tables/{name}/rows", apiHandler.InsertRow)
		r.Put("/tables/{name}/rows", apiHandler.UpdateRow)
		r.Delete("/tables/{name}/rows", apiHandler.DeleteRow)
		r.Post("/query", apiHandler.ExecuteQuery)
	})

	webContent, err := fs.Sub(webFS, "internal/handlers/web")
	if err != nil {
		log.Fatalf("Failed to load web files: %v", err)
	}
	r.Handle("/*", http.FileServer(http.FS(webContent)))

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("\nSQLite Web GUI is running!\n")
	fmt.Printf("Database: %s\n", dbPath)
	fmt.Printf("Open your browser: http://localhost%s\n\n", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
