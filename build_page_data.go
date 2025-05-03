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
	OutputDir       string // e.g. "docs/tech/building_quacker"
	IsIndex         bool
}

type BuildCache struct {
	Pages  []*PageData
	Config *Config
}

// Global var so listpages.go can see it
var globalBuildCache *BuildCache

func assignGlobalCache(cache *BuildCache) {
	globalBuildCache = cache
}

// find .md => "docs/dir/slug" minus "docs/" => "dir/slug"
func FindPageByRelPath(cache *BuildCache, relPath string) string {
	for _, p := range cache.Pages {
		if p.RelPath == relPath {
			if p.OutputDir == "docs" {
				return ""
			}
			if strings.HasPrefix(p.OutputDir, "docs/") {
				prefix := "docs/"
				if len(p.OutputDir) == len(prefix) {
					return ""
				}
				return p.OutputDir[len(prefix):]
			}
			return p.OutputDir
		}
	}
	return ""
}
