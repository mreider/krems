package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// We embed everything under the assets/ directory.
// Make sure your project has an "assets/" folder with
// bootstrap, images, sample markdown, config.yaml, etc.

//go:embed assets/*
var initAssets embed.FS

func handleInit() {
	// Create basic structure
	dirs := []string{
		"markdown",
		"markdown/css",
		"markdown/js",
		"markdown/images",
	}
	for _, d := range dirs {
		err := os.MkdirAll(d, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", d, err)
			os.Exit(1)
		}
	}

	// Copy files from assets/ into the new structure
	err := fs.WalkDir(initAssets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		trimmed := strings.TrimPrefix(path, "assets/")
		srcFile, err := initAssets.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		var destPath string
		switch {
		case strings.HasPrefix(trimmed, "markdown_samples/"):
			// e.g. "markdown_samples/index.md" => "markdown/index.md"
			samplePath := strings.TrimPrefix(trimmed, "markdown_samples/")
			destPath = filepath.Join("markdown", samplePath)
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}
		case strings.HasSuffix(trimmed, ".css") || strings.HasSuffix(trimmed, ".map"):
			destPath = filepath.Join("markdown", "css", filepath.Base(trimmed))
		case strings.HasSuffix(trimmed, ".js"):
			destPath = filepath.Join("markdown", "js", filepath.Base(trimmed))
		case strings.HasSuffix(trimmed, ".png") || strings.HasSuffix(trimmed, ".ico"):
			destPath = filepath.Join("markdown", "images", filepath.Base(trimmed))
		case strings.HasSuffix(trimmed, "config.yaml"):
			destPath = "config.yaml"
		default:
			// Skip anything else
			return nil
		}

		outFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, srcFile)
		if err != nil {
			return err
		}
		fmt.Printf("Created: %s\n", destPath)
		return nil
	})
	if err != nil {
		fmt.Println("Error walking embedded assets:", err)
		os.Exit(1)
	}

	fmt.Println("\nYour sample site structure has been created!")
	fmt.Println("Next steps:")
	fmt.Println("  1) Modify the markdown content in ./markdown")
	fmt.Println("  2) Edit config.yaml as needed (website name, url, menu, etc.)")
	fmt.Println("  3) Run 'krems --build' to generate your static site!")
}
