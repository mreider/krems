package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// handleBuild => krems --build
func handleBuild() {
	// remove docs/ if exists
	_ = os.RemoveAll("docs")

	// read config.yaml
	cfg, err := readConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error reading config.yaml: %v\n", err)
		os.Exit(1)
	}

	// Create internal CSS (Bootstrap, fonts) directly into docs/css
	if err := createInternalCSS("docs"); err != nil {
		fmt.Printf("Error creating internal CSS: %v\n", err)
		os.Exit(1)
	}

	// copy user-provided static assets (js, images) from root => docs/
	if err := copyStaticAssets(); err != nil {
		fmt.Printf("Error copying static assets: %v\n", err)
		os.Exit(1)
	}

	// parse all .md => PageData
	pages, err := parseMarkdownFiles(".") // Changed "markdown" to "."
	if err != nil {
		fmt.Printf("Error parsing markdown: %v\n", err)
		os.Exit(1)
	}

	// create BuildCache
	cache := &BuildCache{
		Pages:  pages,
		Config: cfg,
	}
	assignGlobalCache(cache)

	// process pages => rewrite links => HTML => final
	if err := processPages(cache); err != nil {
		fmt.Printf("Error processing pages: %v\n", err)
		os.Exit(1)
	}

	// Generate author and tag list pages
	if err := generateAuthorPages(cache); err != nil {
		fmt.Printf("Error generating author pages: %v\n", err)
		os.Exit(1)
	}

	if err := generateTagPages(cache); err != nil {
		fmt.Printf("Error generating tag pages: %v\n", err)
		os.Exit(1)
	}

	domain := extractDomain(cache.Config.Website.URL)
	if domain != "" {
		cnameFile := filepath.Join("docs", "CNAME")
		err := os.WriteFile(cnameFile, []byte(domain+"\n"), 0644)
		if err != nil {
			fmt.Printf("Error creating CNAME file: %v\n", err)
			// Decide if you want to exit or just print the error
		} else {
			fmt.Printf("Created: %s (CNAME)\n", cnameFile)
		}
	}

	// generate rss.xml
	if err := generateRSS(cache); err != nil {
		fmt.Printf("Error generating RSS: %v\n", err)
		os.Exit(1)
	}

	// create 404.html
	if err := create404Page(cache); err != nil {
		fmt.Printf("Error creating 404.html: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Build complete! The 'docs/' directory is ready.")
}
