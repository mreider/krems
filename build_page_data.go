package main

import (
	"html/template"
	"strings"
	"time"
)

// PageFrontMatter is the front matter in each .md file
type PageFrontMatter struct {
	Title        string `yaml:"title"`
	Type         string `yaml:"type"` // normal|list
	Description  string `yaml:"description"`
	Image        string `yaml:"image"` // "/images/foo.png" or "images/foo.png"
	Date         string `yaml:"date"`
	ParsedDate   time.Time
	Author       string   `yaml:"author"`
	Tags         []string `yaml:"tags"`
	TagFilter    []string `yaml:"tagFilter"`
	AuthorFilter []string `yaml:"authorFilter"`
}

// PageData captures info for one .md file => HTML page
type PageData struct {
	FrontMatter     PageFrontMatter
	MarkdownContent []byte
	HTMLContent     template.HTML
	RelPath         string // e.g. "tech/Building_Quacker.md"
	OutputDir       string // e.g. "tmp/tech/building_quacker"
	IsIndex         bool
}

type BuildCache struct {
	Pages                 []*PageData
	Config                *Config
	CurrentBuildOutputDir string // Stores the actual output directory for the current build (e.g., "tmp" or a temp path)
}

// Global var so listpages.go can see it
var globalBuildCache *BuildCache

func assignGlobalCache(cache *BuildCache) {
	globalBuildCache = cache
}

// find .md => "outputDir/dir/slug" minus "outputDir/" => "dir/slug"
func FindPageByRelPath(cache *BuildCache, relPath string) string {
	for _, p := range cache.Pages {
		if p.RelPath == relPath {
			// Check if the page's output directory is exactly the build's root output directory
			// (e.g., an index.md at the root of the output)
			if p.OutputDir == cache.CurrentBuildOutputDir {
				// This typically means it's the root index page, so its relative path within the site is effectively empty or handled by BasePath.
				// For constructing links, we might want to return just the slug or an empty string if it's the true site root.
				// Given Krems structure, an OutputDir like "tmp" for an index.md at source root
				// would mean its web path is just "/".
				// If p.OutputDir is "tmp/somepage", we want "somepage".
				// If p.OutputDir is "tmp" (for root index.md), we want "" (empty string) to signify site root.
				return "" // Represents the root of the site if OutputDir matches CurrentBuildOutputDir
			}

			// Construct the prefix to strip, e.g., "tmp/"
			prefixToStrip := cache.CurrentBuildOutputDir + "/"
			if strings.HasPrefix(p.OutputDir, prefixToStrip) {
				// Ensure there's something after the prefix
				if len(p.OutputDir) > len(prefixToStrip) {
					return p.OutputDir[len(prefixToStrip):]
				}
				// If p.OutputDir was "tmp/" (which shouldn't happen if CurrentBuildOutputDir is "tmp"),
				// it would imply a directory named like the output dir itself, treat as root.
				return ""
			}
			// If it doesn't have the prefix (e.g. old scheme or error), return as is, though this case should be rare with new logic.
			return p.OutputDir
		}
	}
	return ""
}
