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
		"markdown/universities",
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
		case strings.HasSuffix(trimmed, ".css"):
			destPath = filepath.Join("markdown", "css", filepath.Base(trimmed))
		case strings.HasSuffix(trimmed, ".woff2"):
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

	// Home index page with list front matter
	indexMD := `---
title: "Krems Home Page"
type: list
---
`
	err = os.WriteFile("markdown/index.md", []byte(indexMD), 0644)
	if err != nil {
		fmt.Printf("Error writing markdown/index.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: markdown/index.md")

	// Krems city info page
	kremsCityMD := `---
title: "Krems City Info"
date: "2024-11-26"
image: "/images/krems1.png"
---
Krems is a beautiful city in Austria known for its rich history, stunning architecture, and vibrant culture.
Explore its winding streets, local markets, and historical landmarks.
`
	err = os.WriteFile("markdown/krems_city_info.md", []byte(kremsCityMD), 0644)
	if err != nil {
		fmt.Printf("Error writing markdown/krems_city_info.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: markdown/krems_city_info.md")

	// Krems travel info page
	kremsTravelMD := `---
title: "Krems Travel Info"
date: "2024-11-26"
image: "/images/krems2.png"
---
Discover the best travel tips and attractions in Krems.
From scenic river walks to local culinary delights, plan your perfect visit to this charming Austrian city.
`
	err = os.WriteFile("markdown/krems_travel_info.md", []byte(kremsTravelMD), 0644)
	if err != nil {
		fmt.Printf("Error writing markdown/krems_travel_info.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: markdown/krems_travel_info.md")

	// Universities subdirectory index page
	univIndexMD := `---
title: "Universities in Krems"
type: list
---
`
	err = os.WriteFile("markdown/universities/index.md", []byte(univIndexMD), 0644)
	if err != nil {
		fmt.Printf("Error writing markdown/universities/index.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: markdown/universities/index.md")

	// University 1 page
	uni1MD := `---
title: "University for Continuing Education Krems"
date: "2024-11-26"
image: "/images/uni1.png"
---
University for Continuing Education Krems is a leading institution in Krems, offering a diverse range of academic programs and cutting-edge research opportunities.
`
	err = os.WriteFile("markdown/universities/uni1.md", []byte(uni1MD), 0644)
	if err != nil {
		fmt.Printf("Error writing markdown/universities/uni1.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: markdown/universities/uni1.md")

	// University 2 page
	uni2MD := `---
title: "IMC Krems University of Applied Sciences"
date: "2024-11-26"
image: "/images/uni2.png"
---
IMC Krems University of Applied Sciences is renowned for its innovative teaching methods and vibrant campus life, making it a hub of academic excellence in Krems.
`
	err = os.WriteFile("markdown/universities/uni2.md", []byte(uni2MD), 0644)
	if err != nil {
		fmt.Printf("Error writing markdown/universities/uni2.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: markdown/universities/uni2.md")

	// Updated config.yaml pointing to the home and universities index pages
	configYAML := `# Config for Krems Static Site
website:
  url: "http://localhost:8080"
  name: "Krems Static Site"

menu:
  - title: "Home"
    path: "index.md"
  - title: "Universities"
    path: "universities/index.md"
`
	if err := os.WriteFile("config.yaml", []byte(configYAML), 0644); err != nil {
		fmt.Printf("Error writing config.yaml: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: config.yaml")
	fmt.Println("\nYour Krems sample site structure has been created!")
	fmt.Println("Next steps:")
	fmt.Println("  1) Modify the markdown content in ./markdown")
	fmt.Println("  2) Edit config.yaml as needed")
	fmt.Println("  3) Run 'krems --build' to generate your static site!")
}
