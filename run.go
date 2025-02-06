package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// handleRun starts a local HTTP server on port 8080
// that serves files from ./docs/ (the output of krems --build).
func handleRun() {
	port := "8080"
	// Wrap Go's FileServer with our custom handler
	// so we can log requests and serve a custom 404 page.
	fs := http.FileServer(http.Dir("docs"))

	http.Handle("/", &loggingFileHandler{
		root:    "docs",
		handler: fs,
	})

	fmt.Printf("Serving 'docs/' on http://localhost:%s ...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// loggingFileHandler intercepts requests to:
//   - Check if path is a directory with no index.html => return 404
//   - If file does not exist => return 404
//   - Otherwise delegate to the FileServer
//
// It also logs each request with the final HTTP status code.
type loggingFileHandler struct {
	root    string // "docs"
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
