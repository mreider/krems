package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"embed" // Added embed here
	"fmt"   // Added fmt here
)

//go:embed assets/css/bootstrap.min.css
var bootstrapCSS embed.FS

//go:embed assets/css/tiempos.woff2
var tiemposFont embed.FS

func createInternalCSS(outputBaseDir string) error {
	cssDir := filepath.Join(outputBaseDir, "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		return fmt.Errorf("failed to create internal css directory %s: %w", cssDir, err)
	}

	// Write bootstrap.min.css
	bootstrapData, err := fs.ReadFile(bootstrapCSS, "assets/css/bootstrap.min.css")
	if err != nil {
		return fmt.Errorf("failed to read embedded bootstrap.min.css: %w", err)
	}
	err = os.WriteFile(filepath.Join(cssDir, "bootstrap.min.css"), bootstrapData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write bootstrap.min.css: %w", err)
	}
	fmt.Printf("Created internal: %s\n", filepath.Join(cssDir, "bootstrap.min.css"))

	// Write tiempos.woff2
	tiemposData, err := fs.ReadFile(tiemposFont, "assets/css/tiempos.woff2")
	if err != nil {
		return fmt.Errorf("failed to read embedded tiempos.woff2: %w", err)
	}
	err = os.WriteFile(filepath.Join(cssDir, "tiempos.woff2"), tiemposData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write tiempos.woff2: %w", err)
	}
	fmt.Printf("Created internal: %s\n", filepath.Join(cssDir, "tiempos.woff2"))

	return nil
}

func copyStaticAssets() error {
	// "css" is removed as it's handled by createInternalCSS
	subdirs := []string{"js", "images"} 
	for _, sd := range subdirs {
		// Source directly from root, e.g., "js", "images"
		src := sd 
		dest := filepath.Join("docs", sd)
		if err := copyDir(src, dest); err != nil {
			// skip if doesn't exist
			var fsErr *fs.PathError
			if errors.Is(err, fs.ErrNotExist) || strings.Contains(err.Error(), "no such file") || errors.As(err, &fsErr) {
				continue
			}
			return err
		}
	}
	return nil
}

func copyDir(src, dest string) error {
	return filepath.Walk(src, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(p, target)
	})
}

func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
