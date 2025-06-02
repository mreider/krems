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
	// Create basic structure at the root
	dirs := []string{
		"universities", // For sample university markdown files
		// "css", // CSS is now handled internally during build
		"js",           // For sample JS (if any, or for user's custom JS)
		"images",       // For sample images
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
			// e.g. "markdown_samples/index.md" => "index.md"
			// e.g. "markdown_samples/universities/index.md" => "universities/index.md"
			samplePath := strings.TrimPrefix(trimmed, "markdown_samples/")
			destPath = samplePath // Place directly at root or in specified subfolder
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}
		// CSS files from embedded assets are no longer copied to root by init;
		// they will be written directly to docs/css during build.
		// case strings.HasSuffix(trimmed, ".css"):
		// 	destPath = filepath.Join("css", filepath.Base(trimmed))
		// case strings.HasSuffix(trimmed, ".woff2"):
		// 	destPath = filepath.Join("css", filepath.Base(trimmed))
		case strings.HasSuffix(trimmed, ".js"):
			destPath = filepath.Join("js", filepath.Base(trimmed)) // If sample JS is provided
		case strings.HasSuffix(trimmed, ".png") || strings.HasSuffix(trimmed, ".ico"):
			destPath = filepath.Join("images", filepath.Base(trimmed))
		case strings.HasSuffix(trimmed, "config.yaml"):
			destPath = "config.yaml" // Stays at the root
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
tagFilter:
  - about
authorFilter:
  - Matt
---
`
	// Note: The embedded assets might already contain an index.md.
	// This explicit write will overwrite it if `assets/markdown_samples/index.md` exists.
	// If the embedded assets are the source of truth for samples, these explicit writes might be redundant
	// or could be removed if the embedded files are correctly placed by the WalkDir logic.
	// For now, keeping them to ensure these specific files are created as per original logic, but paths updated.

	err = os.WriteFile("index.md", []byte(indexMD), 0644)
	if err != nil {
		fmt.Printf("Error writing index.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: index.md")

	// Krems city info page
	kremsCityMD := `---
title: "Krems City Info"
date: "2024-11-26"
image: "/images/krems1.png"
author: "Matt"
tags: ["about"]
---
Krems is a beautiful city in Austria known for its rich history, stunning architecture, and vibrant culture.
Explore its winding streets, local markets, and historical landmarks.
`
	err = os.WriteFile("krems_city_info.md", []byte(kremsCityMD), 0644)
	if err != nil {
		fmt.Printf("Error writing krems_city_info.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: krems_city_info.md")

	// Krems travel info page
	kremsTravelMD := `---
title: "Krems Travel Info"
date: "2024-11-26"
image: "/images/krems2.png"
author: "Matt"
tags: ["about"]
---
Discover the best travel tips and attractions in Krems.
From scenic river walks to local culinary delights, plan your perfect visit to this charming Austrian city.
`
	err = os.WriteFile("krems_travel_info.md", []byte(kremsTravelMD), 0644)
	if err != nil {
		fmt.Printf("Error writing krems_travel_info.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: krems_travel_info.md")

	// Universities subdirectory index page
	univIndexMD := `---
title: "Universities in Krems"
type: list
tagFilter:
  - university
authorFilter:
  - Matt
---
`
	err = os.WriteFile(filepath.Join("universities", "index.md"), []byte(univIndexMD), 0644)
	if err != nil {
		fmt.Printf("Error writing universities/index.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: universities/index.md")

	// University 1 page
	uni1MD := `---
title: "University for Continuing Education Krems"
date: "2024-11-26"
image: "/images/uni1.png"
author: "Matt"
tags: ["university"]
---
University for Continuing Education Krems is a leading institution in Krems, offering a diverse range of academic programs and cutting-edge research opportunities.
`
	err = os.WriteFile(filepath.Join("universities", "uni1.md"), []byte(uni1MD), 0644)
	if err != nil {
		fmt.Printf("Error writing universities/uni1.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: universities/uni1.md")

	// University 2 page
	uni2MD := `---
title: "IMC Krems University of Applied Sciences"
date: "2024-11-26"
image: "/images/uni2.png"
author: "Matt"
tags: ["university"]
---
IMC Krems University of Applied Sciences is renowned for its innovative teaching methods and vibrant campus life, making it a hub of academic excellence in Krems.
`
	err = os.WriteFile(filepath.Join("universities", "uni2.md"), []byte(uni2MD), 0644)
	if err != nil {
		fmt.Printf("Error writing universities/uni2.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: universities/uni2.md")

	// Updated config.yaml pointing to the home and universities index pages
	// This will also be created by the WalkDir if assets/config.yaml exists.
	// If embedded config is preferred, this explicit write can be removed.
	configYAML := `# Config for Krems Static Site
website:
  url: "http://localhost:8080" # Default, user should change this
  name: "My Krems Site"       # Default, user should change this

menu:
  - title: "Home"
    path: "index.md" # Relative to root
  - title: "Universities"
    path: "universities/index.md" # Relative to root
  - title: "City Info"
    path: "krems_city_info.md"
  - title: "Travel Info"
    path: "krems_travel_info.md"

`
	// This ensures the config.yaml is created/overwritten with these specific menu paths.
	// If an assets/config.yaml is embedded, it might be copied first by WalkDir, then overwritten here.
	if err := os.WriteFile("config.yaml", []byte(configYAML), 0644); err != nil {
		fmt.Printf("Error writing config.yaml: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created: config.yaml") // This might print twice if WalkDir also copies it.

	fmt.Println("\nYour Krems sample site structure has been created at the root level!")
	fmt.Println("Next steps:")
	fmt.Println("  1) Modify the markdown content in the current directory (.) and its subfolders (e.g., ./universities).")
	fmt.Println("  2) Edit config.yaml as needed (e.g., update website.url, website.name).")
	fmt.Println("  3) Run 'krems --build' to generate your static site into the 'docs/' folder.")
	fmt.Println("  4) Run 'krems --run' to preview your site locally at http://localhost:8080.")
}
