package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// handleBuild => krems --build
// isDevMode indicates if the build is for local development (krems --run)
// outputDir specifies where to build the site.
func handleBuild(isDevMode bool, outputDir string) {
	// remove outputDir if exists
	_ = os.RemoveAll(outputDir)

	// read config.yaml
	cfg, err := readConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error reading config.yaml: %v\n", err)
		os.Exit(1)
	}

	// Determine the effective base path
	// If isDevMode is true and DevPath is set, use DevPath. Otherwise, use BasePath.
	if isDevMode && cfg.Website.DevPath != "" {
		cfg.Website.BasePath = cfg.Website.DevPath
	}
	// If not in dev mode, or DevPath is not set, cfg.Website.BasePath remains as read from config.yaml
	// or its default if not specified.

	// Create internal CSS (Bootstrap, fonts) directly into outputDir/css
	if err := createInternalCSS(outputDir); err != nil { // MODIFIED
		fmt.Printf("Error creating internal CSS: %v\n", err)
		os.Exit(1)
	}

	// Create internal JS (Bootstrap) directly into outputDir/js
	if err := createInternalJS(outputDir); err != nil { // MODIFIED
		fmt.Printf("Error creating internal JS: %v\n", err)
		os.Exit(1)
	}

	// Create internal favicon directly into outputDir/images
	if err := createInternalFavicon(outputDir); err != nil { // MODIFIED
		fmt.Printf("Error creating internal favicon: %v\n", err)
		os.Exit(1)
	}

	// copy user-provided static assets (js, images) from root => outputDir/
	// This will overwrite embedded files if user provides their own versions.
	if err := copyStaticAssets(outputDir); err != nil { // MODIFIED (assuming copyStaticAssets will take outputDir)
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
		Pages:                 pages,
		Config:                cfg,
		CurrentBuildOutputDir: outputDir, // Set the current build output directory
	}
	assignGlobalCache(cache)

	// process pages => rewrite links => HTML => final
	if err := processPages(cache, outputDir); err != nil { // MODIFIED (assuming processPages will take outputDir)
		fmt.Printf("Error processing pages: %v\n", err)
		os.Exit(1)
	}

	// Generate author and tag list pages
	if err := generateAuthorPages(cache, outputDir); err != nil { // MODIFIED (assuming generateAuthorPages will take outputDir)
		fmt.Printf("Error generating author pages: %v\n", err)
		os.Exit(1)
	}

	if err := generateTagPages(cache, outputDir); err != nil { // MODIFIED (assuming generateTagPages will take outputDir)
		fmt.Printf("Error generating tag pages: %v\n", err)
		os.Exit(1)
	}

	domain := extractDomain(cache.Config.Website.URL)
	if domain != "" {
		cnameFile := filepath.Join(outputDir, "CNAME") // MODIFIED
		err := os.WriteFile(cnameFile, []byte(domain+"\n"), 0644)
		if err != nil {
			fmt.Printf("Error creating CNAME file: %v\n", err)
			// Decide if you want to exit or just print the error
		} else {
			fmt.Printf("Created: %s (CNAME)\n", cnameFile)
		}
	}

	// generate rss.xml
	if err := generateRSS(cache, outputDir); err != nil { // MODIFIED (assuming generateRSS will take outputDir)
		fmt.Printf("Error generating RSS: %v\n", err)
		os.Exit(1)
	}

	// create 404.html
	if err := create404Page(cache, outputDir); err != nil { // MODIFIED (assuming create404Page will take outputDir)
		fmt.Printf("Error creating 404.html: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Build complete! The '%s' directory is ready.\n", outputDir) // MODIFIED
}
