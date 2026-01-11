package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rzhade3/sqlite-webgui/internal/database"
	"github.com/rzhade3/sqlite-webgui/internal/handlers"
)

//go:embed internal/handlers/web/*
var webFS embed.FS

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // linux, freebsd, openbsd, netbsd
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	writable := flag.Bool("writable", false, "Enable write operations (default: false, read-only mode)")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--port PORT] [--writable] <database.db>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fmt.Fprintf(os.Stderr, "  --port PORT    Port to run the server on (default: 8080)\n")
		fmt.Fprintf(os.Stderr, "  --writable     Enable write operations (default: read-only mode)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s mydata.db                  # Read-only mode (safe)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --writable mydata.db       # Enable write operations\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --port 3000 mydata.db      # Custom port, read-only\n", os.Args[0])
		os.Exit(1)
	}

	dbPath := args[0]

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatalf("Database file does not exist: %s", dbPath)
	}

	readonly := !*writable
	db, err := database.New(dbPath, readonly)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	apiHandler := handlers.NewAPIHandler(db)

	r.Route("/api", func(r chi.Router) {
		r.Get("/mode", apiHandler.GetMode)
		r.Get("/tables", apiHandler.GetTables)
		r.Get("/tables/{name}/schema", apiHandler.GetTableSchema)
		r.Get("/tables/{name}/data", apiHandler.GetTableData)
		r.Post("/query", apiHandler.ExecuteQuery)

		// Only register write endpoints if database is not in read-only mode
		if !db.IsReadOnly() {
			r.Post("/tables/{name}/rows", apiHandler.InsertRow)
			r.Put("/tables/{name}/rows", apiHandler.UpdateRow)
			r.Delete("/tables/{name}/rows", apiHandler.DeleteRow)
		}
	})

	webContent, err := fs.Sub(webFS, "internal/handlers/web")
	if err != nil {
		log.Fatalf("Failed to load web files: %v", err)
	}
	r.Handle("/*", http.FileServer(http.FS(webContent)))

	addr := fmt.Sprintf(":%s", *port)
	url := fmt.Sprintf("http://localhost%s", addr)
	fmt.Printf("\nSQLite Web GUI is running!\n")
	fmt.Printf("Database: %s\n", dbPath)
	if *writable {
		fmt.Printf("Mode: READ-WRITE\n")
	} else {
		fmt.Printf("Mode: READ-ONLY\n")
	}
	fmt.Printf("Open your browser: %s\n\n", url)

	if err := openBrowser(url); err != nil {
		log.Printf("Failed to open browser automatically: %v", err)
	}

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
