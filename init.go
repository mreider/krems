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
		"markdown/topics",
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

	indexMD := `---
title: "All About Mollusks"
image: "images/mollusk.png"
---

Welcome to our sample site about mollusks! Just a short piece of text here
about snails and slugs and other interesting creatures. 
`
	err = os.WriteFile("markdown/index.md", []byte(indexMD), 0644)
    if err != nil {
        fmt.Printf("Error writing markdown/index.md: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Created: markdown/index.md")

    // Subdirectory page: topics/index.md
    topicsMD := `---
title: "Mollusk Topics"
image: "images/mollusk2.png"
---

Here we discuss more detailed topics related to mollusks in a subdirectory.
Short text again for demonstration.
`
    err = os.WriteFile("markdown/topics/index.md", []byte(topicsMD), 0644)
    if err != nil {
        fmt.Printf("Error writing markdown/topics/index.md: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Created: markdown/topics/index.md")


    configYAML := `# Example config for local test site
website:
  url: "http://localhost:8080"
  name: "Local Test Site"

menu:
  - title: "Home"
    path: "index.md"
  - title: "Topics"
    path: "topics/index.md"
`
    if err := os.WriteFile("config.yaml", []byte(configYAML), 0644); err != nil {
        fmt.Printf("Error writing config.yaml: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Created: config.yaml")
    fmt.Println("\nYour sample site structure has been created!")
    fmt.Println("Next steps:")
    fmt.Println("  1) Modify the markdown content in ./markdown")
    fmt.Println("  2) Edit config.yaml as needed")
    fmt.Println("  3) Run 'krems --build' to generate your static site!")
}
