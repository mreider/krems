package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/gosimple/slug"
)

// generateAuthorPages generates list pages for each author
func generateAuthorPages(cache *BuildCache, outputDirRoot string) error { // MODIFIED: Added outputDirRoot
	authors := make(map[string]bool)
	for _, p := range cache.Pages {
		if p.FrontMatter.Author != "" {
			authors[p.FrontMatter.Author] = true
		}
	}

	for author := range authors {
		if err := generateAuthorPage(cache, author, outputDirRoot); err != nil { // MODIFIED: Passed outputDirRoot
			return err
		}
	}
	return nil
}

// generateTagPages generates list pages for each tag
func generateTagPages(cache *BuildCache, outputDirRoot string) error { // MODIFIED: Added outputDirRoot
	tags := make(map[string]bool)
	for _, p := range cache.Pages {
		for _, tag := range p.FrontMatter.Tags {
			tags[tag] = true
		}
	}

	for tag := range tags {
		if err := generateTagPage(cache, tag, outputDirRoot); err != nil { // MODIFIED: Passed outputDirRoot
			return err
		}
	}
	return nil
}

// generateAuthorPage generates a list page for a specific author
func generateAuthorPage(cache *BuildCache, author string, outputDirRoot string) error { // MODIFIED: Added outputDirRoot
	authorSlug := slug.Make(author)
	dir := filepath.Join(outputDirRoot, "authors", authorSlug) // MODIFIED: Used outputDirRoot
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	outFile := filepath.Join(dir, "index.html")
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a unique RelPath for this author page
	authorRelPath := filepath.Join("authors", authorSlug, "index.md")
	
	pseudo := &PageData{
		FrontMatter: PageFrontMatter{
			Title:        fmt.Sprintf("Posts by %s", author),
			Type:         "list",
			AuthorFilter: []string{author}, // Filter by the current author
		},
		RelPath:   authorRelPath,
		OutputDir: dir,
	}
	
	// Add the pseudo page to the cache so listPagesInDirectory can find it
	cache.Pages = append(cache.Pages, pseudo)

	var menuItems []string
	var menuTargets []string
	for _, item := range cache.Config.Menu {
		menuItems = append(menuItems, item.Title)
		outDir := FindPageByRelPath(cache, item.Path)
		if outDir == "" {
			if item.Path == "index.md" {
				menuTargets = append(menuTargets, "/")
			} else {
				menuTargets = append(menuTargets, "/")
			}
		} else {
			menuTargets = append(menuTargets, "/"+outDir+"/")
		}
	}

	data := struct {
		Config      *Config
		Page        *PageData
		MenuItems   []string
		MenuTargets []string
	}{
		Config:      cache.Config,
		Page:        pseudo,
		MenuItems:   menuItems,
		MenuTargets: menuTargets,
	}

	tmpl := template.New("author")
	tmpl = initTemplateFuncs(tmpl, cache, outputDirRoot) // MODIFIED: Passed cache and outputDirRoot as siteBuildRoot
	tmpl, err = tmpl.Parse(htmlTemplate)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	fmt.Printf("Generated: %s\n", outFile)
	return nil
}

// generateTagPage generates a list page for a specific tag
func generateTagPage(cache *BuildCache, tag string, outputDirRoot string) error { // MODIFIED: Added outputDirRoot
	tagSlug := slug.Make(tag)
	dir := filepath.Join(outputDirRoot, "tags", tagSlug) // MODIFIED: Used outputDirRoot
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	outFile := filepath.Join(dir, "index.html")
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a unique RelPath for this tag page
	tagRelPath := filepath.Join("tags", tagSlug, "index.md")
	
	pseudo := &PageData{
		FrontMatter: PageFrontMatter{
			Title:     fmt.Sprintf("Posts tagged with %s", tag),
			Type:      "list",
			TagFilter: []string{tag}, // Filter by the current tag
		},
		RelPath:   tagRelPath,
		OutputDir: dir,
	}
	
	// Add the pseudo page to the cache so listPagesInDirectory can find it
	cache.Pages = append(cache.Pages, pseudo)

	data := struct {
		Config      *Config
		Page        *PageData
		MenuItems   []string
		MenuTargets []string
	}{
		Config:      cache.Config,
		Page:        pseudo,
		MenuItems:   []string{},
		MenuTargets: []string{},
	}

	var menuItems []string
	var menuTargets []string
	for _, item := range cache.Config.Menu {
		menuItems = append(menuItems, item.Title)
		outDir := FindPageByRelPath(cache, item.Path)
		if outDir == "" {
			if item.Path == "index.md" {
				menuTargets = append(menuTargets, "/")
			} else {
				menuTargets = append(menuTargets, "/")
			}
		} else {
			menuTargets = append(menuTargets, "/"+outDir+"/")
		}
	}

	data.MenuItems = menuItems
	data.MenuTargets = menuTargets

	tmpl := template.New("tag")
	tmpl = initTemplateFuncs(tmpl, cache, outputDirRoot) // MODIFIED: Passed cache and outputDirRoot as siteBuildRoot
	tmpl, err = tmpl.Parse(htmlTemplate)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	fmt.Printf("Generated: %s\n", outFile)
	return nil
}
