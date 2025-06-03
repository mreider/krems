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

// handleRun builds the site into a temporary directory,
// starts a local HTTP server to serve it, and cleans up the directory on exit.
func handleRun() {
	// Create a temporary directory for the build output
	tempDir, err := os.MkdirTemp("", "krems-run-")
	if err != nil {
		log.Fatalf("Failed to create temporary directory: %v", err)
	}
	fmt.Printf("Using temporary directory for build: %s\n", tempDir)

	// Defer cleanup of the temporary directory
	// Also set up a signal handler for Ctrl+C
	cleanup := func() {
		fmt.Printf("\nCleaning up temporary directory: %s\n", tempDir)
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("Warning: Failed to remove temporary directory %s: %v", tempDir, err)
		}
	}
	defer cleanup()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(0)
	}()

	// Build the site into the temporary directory
	// Assuming handleBuild will be modified to accept an output directory
	// and a flag indicating it's for 'run' mode (which might affect base paths etc.)
	fmt.Println("Building site for local preview...")
	handleBuild(true, tempDir) // true for isDevMode, tempDir for output
	fmt.Println("Build complete.")

	port := "8080"
	fs := http.FileServer(http.Dir(tempDir))

	http.Handle("/", &loggingFileHandler{
		root:    tempDir,
		handler: fs,
	})

	fmt.Printf("Serving '%s' on http://localhost:%s ... (Press Ctrl+C to stop)\n", tempDir, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// loggingFileHandler intercepts requests to:
//   - Check if path is a directory with no index.html => return 404
//   - If file does not exist => return 404
//   - Otherwise delegate to the FileServer
//
// It also logs each request with the final HTTP status code.
type loggingFileHandler struct {
	root    string // temporary build directory
	handler http.Handler
}

func (l *loggingFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}

	fullPath := filepath.Join(l.root, r.URL.Path)
	info, err := os.Stat(fullPath)
	if err != nil {
		// File does not exist => 404
		l.notFound(lrw, r)
		return
	}

	if info.IsDir() {
		// Check if there's an index.html in that directory
		indexPath := filepath.Join(fullPath, "index.html")
		fi, err2 := os.Stat(indexPath)
		if err2 != nil || fi.IsDir() {
			// No index.html => 404
			l.notFound(lrw, r)
			return
		}
		// Otherwise, let FileServer serve the directory (which includes index.html).
	}

	// If it's a file, or a directory with index.html => pass to FileServer
	l.handler.ServeHTTP(lrw, r)
	log.Printf("%s %s -> %d\n", r.Method, r.URL.Path, lrw.statusCode)
}

// notFound is a helper that writes a 404 status and attempts
// to serve docs/404.html if present. If not present,
// it serves a minimal fallback HTML.
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

// loggingResponseWriter records the final status code written,
// so we can log it after FileServer (or our custom logic) finishes.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
