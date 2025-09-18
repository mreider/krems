package main

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"embed"
	"fmt"
	"text/template"
)

// Embed assets from a flat assets/ directory
//go:embed assets/bootstrap.min.css
var embeddedBootstrapCSS embed.FS

//go:embed assets/lora-regular.woff2
var embeddedLoraRegularFont embed.FS

//go:embed assets/lora-italic.woff2
var embeddedLoraItalicFont embed.FS

//go:embed assets/source-sans-regular.woff2
var embeddedSourceSansFont embed.FS

//go:embed assets/bootstrap.js
var embeddedBootstrapJS embed.FS

//go:embed assets/favicon.ico
var embeddedFaviconICO embed.FS

//go:embed assets/custom.css
var embeddedCustomCSS embed.FS

func createInternalCSS(outputBaseDir string) error {
	cssDir := filepath.Join(outputBaseDir, "css")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		return fmt.Errorf("failed to create output css directory %s: %w", cssDir, err)
	}

	bootstrapData, err := fs.ReadFile(embeddedBootstrapCSS, "assets/bootstrap.min.css")
	if err != nil {
		return fmt.Errorf("failed to read embedded bootstrap.min.css: %w", err)
	}
	err = os.WriteFile(filepath.Join(cssDir, "bootstrap.min.css"), bootstrapData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write bootstrap.min.css: %w", err)
	}
	fmt.Printf("Created internal: %s\n", filepath.Join(cssDir, "bootstrap.min.css"))

	// Lora regular font
	loraRegularData, err := fs.ReadFile(embeddedLoraRegularFont, "assets/lora-regular.woff2")
	if err != nil {
		return fmt.Errorf("failed to read embedded lora-regular.woff2: %w", err)
	}
	err = os.WriteFile(filepath.Join(cssDir, "lora-regular.woff2"), loraRegularData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write lora-regular.woff2: %w", err)
	}
	fmt.Printf("Created internal: %s\n", filepath.Join(cssDir, "lora-regular.woff2"))

	// Lora italic font
	loraItalicData, err := fs.ReadFile(embeddedLoraItalicFont, "assets/lora-italic.woff2")
	if err != nil {
		return fmt.Errorf("failed to read embedded lora-italic.woff2: %w", err)
	}
	err = os.WriteFile(filepath.Join(cssDir, "lora-italic.woff2"), loraItalicData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write lora-italic.woff2: %w", err)
	}
	fmt.Printf("Created internal: %s\n", filepath.Join(cssDir, "lora-italic.woff2"))

	// Source Sans font
	sourceSansData, err := fs.ReadFile(embeddedSourceSansFont, "assets/source-sans-regular.woff2")
	if err != nil {
		return fmt.Errorf("failed to read embedded source-sans-regular.woff2: %w", err)
	}
	err = os.WriteFile(filepath.Join(cssDir, "source-sans-regular.woff2"), sourceSansData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write source-sans-regular.woff2: %w", err)
	}
	fmt.Printf("Created internal: %s\n", filepath.Join(cssDir, "source-sans-regular.woff2"))

	customCSSData, err := fs.ReadFile(embeddedCustomCSS, "assets/custom.css")
	if err != nil {
		return fmt.Errorf("failed to read embedded custom.css: %w", err)
	}

	// Process CSS through template engine to resolve sitePath functions
	processedCSS, err := processCSSThroughTemplate(string(customCSSData))
	if err != nil {
		return fmt.Errorf("failed to process custom.css through template: %w", err)
	}

	err = os.WriteFile(filepath.Join(cssDir, "custom.css"), []byte(processedCSS), 0644)
	if err != nil {
		return fmt.Errorf("failed to write custom.css: %w", err)
	}
	fmt.Printf("Created internal: %s\n", filepath.Join(cssDir, "custom.css"))

	return nil
}

func createInternalJS(outputBaseDir string) error {
	jsDir := filepath.Join(outputBaseDir, "js")
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		return fmt.Errorf("failed to create output js directory %s: %w", jsDir, err)
	}

	jsData, err := fs.ReadFile(embeddedBootstrapJS, "assets/bootstrap.js")
	if err != nil {
		return fmt.Errorf("failed to read embedded bootstrap.js: %w", err)
	}
	err = os.WriteFile(filepath.Join(jsDir, "bootstrap.js"), jsData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write bootstrap.js: %w", err)
	}
	fmt.Printf("Created internal: %s\n", filepath.Join(jsDir, "bootstrap.js"))
	return nil
}

func createInternalFavicon(outputBaseDir string) error {
	// Favicon goes into outputBaseDir/images/favicon.ico, but is sourced from assets/favicon.ico
	// The HTML template links to {{sitePath "/images/favicon.ico"}}
	imagesDir := filepath.Join(outputBaseDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create output images directory %s: %w", imagesDir, err)
	}

	faviconData, err := fs.ReadFile(embeddedFaviconICO, "assets/favicon.ico")
	if err != nil {
		return fmt.Errorf("failed to read embedded favicon.ico: %w", err)
	}
	err = os.WriteFile(filepath.Join(imagesDir, "favicon.ico"), faviconData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write favicon.ico: %w", err)
	}
	fmt.Printf("Created internal: %s\n", filepath.Join(imagesDir, "favicon.ico"))
	return nil
}

func copyStaticAssets(outputDir string) error {
	// "css" is removed as it's handled by createInternalCSS
	subdirs := []string{"js", "images"}
	for _, sd := range subdirs {
		// Source directly from root, e.g., "js", "images"
		src := sd
		dest := filepath.Join(outputDir, sd)
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

// processCSSThroughTemplate processes CSS content through the template engine
// to resolve sitePath and other template functions
func processCSSThroughTemplate(cssContent string) (string, error) {
	// Create a new template with the same functions as the HTML template
	tmpl := template.New("css").Funcs(template.FuncMap{
		"sitePath": sitePath,
	})

	// Parse the CSS as a template
	_, err := tmpl.Parse(cssContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse CSS template: %w", err)
	}

	// Execute the template to resolve the sitePath functions
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		return "", fmt.Errorf("failed to execute CSS template: %w", err)
	}

	return buf.String(), nil
}
