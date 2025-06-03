package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// handleRun builds the site into the ./.tmp directory,
// starts a local HTTP server to serve it, and cleans up the directory on exit.
func handleRun(port string) { // Accept port as a parameter
	// Use the constant outputDirName = ".tmp"
	// This directory is relative to where krems is run (project root)
	
	// Ensure the output directory exists, remove if it does to start fresh for run
	// For 'run', we always want a fresh build in .tmp
	if err := os.RemoveAll(outputDirName); err != nil {
		log.Printf("Warning: Failed to remove existing %s directory: %v", outputDirName, err)
		// Continue, as handleBuild will also try to remove it.
	}
	if err := os.MkdirAll(outputDirName, 0755); err != nil {
		log.Fatalf("Failed to create %s directory: %v", outputDirName, err)
	}
	fmt.Printf("Using output directory for build: %s\n", outputDirName)

	// Defer cleanup of the .tmp directory
	// Also set up a signal handler for Ctrl+C
	cleanup := func() {
		fmt.Printf("\nCleaning up output directory: %s\n", outputDirName)
		if err := os.RemoveAll(outputDirName); err != nil {
			log.Printf("Warning: Failed to remove output directory %s: %v", outputDirName, err)
		}
	}
	defer cleanup()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup() // Ensure cleanup happens before exit
		os.Exit(0)
	}()

	fmt.Println("Building site for local preview...")
	handleBuild(true, outputDirName) // true for isDevMode, outputDirName for output
	fmt.Println("Build complete.")

	// Use the port parameter
	fs := http.FileServer(http.Dir(outputDirName))

	http.Handle("/", &loggingFileHandler{
		root:    outputDirName, // Use the .tmp directory
		handler: fs,
	})

	fmt.Printf("Serving '%s' on http://localhost:%s ... (Press Ctrl+C to stop)\n", outputDirName, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// loggingFileHandler intercepts requests to:
//   - Check if path is a directory with no index.html => return 404
//   - If file does not exist => return 404
//   - Otherwise delegate to the FileServer
//
// It also logs each request with the final HTTP status code.
type loggingFileHandler struct {
	root    string // e.g. ".tmp"
	handler http.Handler
}

func (l *loggingFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}

	fullPath := filepath.Join(l.root, r.URL.Path)
	info, err := os.Stat(fullPath)
	if err != nil {
		l.notFound(lrw, r)
		return
	}

	if info.IsDir() {
		indexPath := filepath.Join(fullPath, "index.html")
		fi, err2 := os.Stat(indexPath)
		if err2 != nil || fi.IsDir() {
			l.notFound(lrw, r)
			return
		}
	}

	l.handler.ServeHTTP(lrw, r)
	log.Printf("%s %s -> %d\n", r.Method, r.URL.Path, lrw.statusCode)
}

func (l *loggingFileHandler) notFound(lrw *loggingResponseWriter, r *http.Request) {
	lrw.WriteHeader(http.StatusNotFound)

	notFoundPath := filepath.Join(l.root, "404.html")
	if _, err := os.Stat(notFoundPath); err == nil {
		http.ServeFile(lrw, r, notFoundPath)
	} else {
		fmt.Fprintln(lrw, "<html><body><h1>404 Not Found</h1></body></html>")
	}
	log.Printf("%s %s -> %d\n", r.Method, r.URL.Path, http.StatusNotFound)
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
