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

	// Handle CSS
	if cfg.Website.AlternativeCSSDir != "" {
		fmt.Printf("Using alternative CSS from: %s\n", cfg.Website.AlternativeCSSDir)
		cssOutputDir := filepath.Join(outputDir, "css")
		if err := os.MkdirAll(cssOutputDir, 0755); err != nil {
			fmt.Printf("Error creating css output directory %s: %v\n", cssOutputDir, err)
			os.Exit(1)
		}
		files, err := os.ReadDir(cfg.Website.AlternativeCSSDir)
		if err != nil {
			fmt.Printf("Error reading alternative CSS directory %s: %v\n", cfg.Website.AlternativeCSSDir, err)
			os.Exit(1)
		}
		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".css" {
				srcPath := filepath.Join(cfg.Website.AlternativeCSSDir, file.Name())
				destPath := filepath.Join(cssOutputDir, file.Name())
				if err := copyFile(srcPath, destPath); err != nil {
					fmt.Printf("Error copying alternative CSS file %s to %s: %v\n", srcPath, destPath, err)
					os.Exit(1)
				}
				fmt.Printf("Copied alternative CSS: %s\n", destPath)
			}
		}
	} else {
		if err := createInternalCSS(outputDir); err != nil {
			fmt.Printf("Error creating internal CSS: %v\n", err)
			os.Exit(1)
		}
	}

	// Handle JS
	if cfg.Website.AlternativeJSDir != "" {
		fmt.Printf("Using alternative JS from: %s\n", cfg.Website.AlternativeJSDir)
		jsOutputDir := filepath.Join(outputDir, "js")
		if err := os.MkdirAll(jsOutputDir, 0755); err != nil {
			fmt.Printf("Error creating js output directory %s: %v\n", jsOutputDir, err)
			os.Exit(1)
		}
		files, err := os.ReadDir(cfg.Website.AlternativeJSDir)
		if err != nil {
			fmt.Printf("Error reading alternative JS directory %s: %v\n", cfg.Website.AlternativeJSDir, err)
			os.Exit(1)
		}
		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".js" {
				srcPath := filepath.Join(cfg.Website.AlternativeJSDir, file.Name())
				destPath := filepath.Join(jsOutputDir, file.Name())
				if err := copyFile(srcPath, destPath); err != nil {
					fmt.Printf("Error copying alternative JS file %s to %s: %v\n", srcPath, destPath, err)
					os.Exit(1)
				}
				fmt.Printf("Copied alternative JS: %s\n", destPath)
			}
		}
	} else {
		if err := createInternalJS(outputDir); err != nil {
			fmt.Printf("Error creating internal JS: %v\n", err)
			os.Exit(1)
		}
	}

	// Handle Favicon
	if cfg.Website.AlternativeFavicon != "" {
		fmt.Printf("Using alternative favicon from: %s\n", cfg.Website.AlternativeFavicon)
		imagesOutputDir := filepath.Join(outputDir, "images")
		if err := os.MkdirAll(imagesOutputDir, 0755); err != nil {
			fmt.Printf("Error creating images output directory %s: %v\n", imagesOutputDir, err)
			os.Exit(1)
		}
		// Assuming the alternative favicon should be named favicon.ico in the output
		destPath := filepath.Join(imagesOutputDir, "favicon.ico")
		if err := copyFile(cfg.Website.AlternativeFavicon, destPath); err != nil {
			fmt.Printf("Error copying alternative favicon from %s to %s: %v\n", cfg.Website.AlternativeFavicon, destPath, err)
			os.Exit(1)
		}
		fmt.Printf("Copied alternative favicon: %s\n", destPath)
	} else {
		if err := createInternalFavicon(outputDir); err != nil {
			fmt.Printf("Error creating internal favicon: %v\n", err)
			os.Exit(1)
		}
	}

	// copy user-provided static assets (js, images) from root => outputDir/
	// This will overwrite embedded files if user provides their own versions.
	if err := copyStaticAssets(outputDir); err != nil {
		fmt.Printf("Error copying static assets: %v\n", err)
		os.Exit(1)
	}

	// parse all .md => PageData
	pages, err := parseMarkdownFiles(".")
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
	if err := processPages(cache, outputDir); err != nil {
		fmt.Printf("Error processing pages: %v\n", err)
		os.Exit(1)
	}

	// Generate author and tag list pages
	if err := generateAuthorPages(cache, outputDir); err != nil {
		fmt.Printf("Error generating author pages: %v\n", err)
		os.Exit(1)
	}

	if err := generateTagPages(cache, outputDir); err != nil {
		fmt.Printf("Error generating tag pages: %v\n", err)
		os.Exit(1)
	}

	domain := extractDomain(cache.Config.Website.URL)
	if domain != "" {
		cnameFile := filepath.Join(outputDir, "CNAME")
		err := os.WriteFile(cnameFile, []byte(domain+"\n"), 0644)
		if err != nil {
			fmt.Printf("Error creating CNAME file: %v\n", err)
			// Decide if you want to exit or just print the error
		} else {
			fmt.Printf("Created: %s (CNAME)\n", cnameFile)
		}
	}

	// generate rss.xml
	if err := generateRSS(cache, outputDir); err != nil {
		fmt.Printf("Error generating RSS: %v\n", err)
		os.Exit(1)
	}

	// create 404.html
	if err := create404Page(cache, outputDir); err != nil {
		fmt.Printf("Error creating 404.html: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Build complete! The '%s' directory is ready.\n", outputDir)
}
