package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// parse markdown => PageData
func parseMarkdownFiles(root string) ([]*PageData, error) {
	var pages []*PageData
	err := filepath.Walk(root, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(p), ".md") {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		page, err := parseFrontMatter(raw)
		if err != nil {
			return fmt.Errorf("error parsing front matter in %s: %w", p, err)
		}
		page.RelPath = filepath.ToSlash(rel)
		pages = append(pages, page)
		return nil
	})
	return pages, err
}

func parseFrontMatter(fileBytes []byte) (*PageData, error) {
	page := &PageData{}
	delim := []byte("---")
	parts := bytes.SplitN(fileBytes, delim, 3)
	if len(parts) == 3 {
		fmBytes := bytes.TrimSpace(parts[1])
		var fm PageFrontMatter
		if err := yaml.Unmarshal(fmBytes, &fm); err != nil {
			return nil, err
		}
		if fm.Type == "" {
			fm.Type = "normal"
		}
		if fm.Date != "" {
			if t, err := time.Parse("2006-01-02", fm.Date); err == nil {
				fm.ParsedDate = t
			}
		}
		page.FrontMatter = fm
		page.MarkdownContent = bytes.TrimSpace(parts[2])
	} else {
		page.FrontMatter = PageFrontMatter{Type: "normal"}
		page.MarkdownContent = fileBytes
	}
	return page, nil
}

// In build.go, modify the fixLinksAndImages function to better handle "../" relative paths
func fixLinksAndImages(cache *BuildCache, page *PageData) []byte {
	lines := bytes.Split(page.MarkdownContent, []byte("\n"))
	reLink := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	reImg := regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)

	for i, line := range lines {
			// images => 400px
			line = reImg.ReplaceAllFunc(line, func(m []byte) []byte {
					sub := reImg.FindSubmatch(m)
					if len(sub) < 3 {
							return m
					}
					alt := string(sub[1])
					imgPath := string(sub[2])
					return []byte(fmt.Sprintf(
							`<img src="%s" alt="%s" style="max-width:800px;width:100%%;height:auto;" class="mb-3 img-fluid border border-1 border-dark"/>`,
							imgPath, alt))
			})

			// local .md => /slug/
			line = reLink.ReplaceAllFunc(line, func(m []byte) []byte {
					sub := reLink.FindSubmatch(m)
					if len(sub) < 3 {
							return m
					}
					linkText := string(sub[1])
					linkTarget := string(sub[2])
					lc := strings.ToLower(linkTarget)

					if strings.HasPrefix(lc, "http://") || strings.HasPrefix(lc, "https://") {
							return m // external => no rewrite
					}
					
					if strings.HasSuffix(lc, ".md") {
							// Handle relative paths starting with ../
							linkCandidate := linkTarget
							if strings.HasPrefix(linkCandidate, "../") {
									// Get the current page's directory path
									dirOfPage := filepath.Dir(page.RelPath)
									if dirOfPage == "." {
											dirOfPage = ""
									}
									
									// Resolve the relative path properly
									resolvedPath := filepath.Clean(filepath.Join(dirOfPage, linkCandidate))
									linkCandidate = filepath.ToSlash(resolvedPath)
							} else if !strings.Contains(linkCandidate, "/") {
									// If no slash, assume same dir => join with page's dir
									dirOfPage := filepath.Dir(page.RelPath)
									if dirOfPage == "." {
											dirOfPage = ""
									}
									linkCandidate = filepath.Join(dirOfPage, linkCandidate)
							}
							
							linkCandidate = filepath.ToSlash(linkCandidate)

							// find a matching .RelPath
							for _, other := range cache.Pages {
									if other.RelPath == linkCandidate {
											outDir := FindPageByRelPath(cache, other.RelPath)
											return []byte(fmt.Sprintf("[%s](/%s/)", linkText, outDir))
									}
							}
					}
					return m
			})

			lines[i] = line
	}
	return bytes.Join(lines, []byte("\n"))
}