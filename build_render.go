package main

import (
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gosimple/slug"
)

func processPages(cache *BuildCache) error {
	for _, p := range cache.Pages {
		base := filepath.Base(p.RelPath)
		if base == "index.md" {
			p.IsIndex = true
		} else {
			p.IsIndex = false
		}
		dir := filepath.Dir(p.RelPath)
		if dir == "." {
			dir = ""
		}

		if p.IsIndex {
			p.OutputDir = filepath.Join("docs", dir)
		} else {
			slugTitle := slug.Make(p.FrontMatter.Title)
			if slugTitle == "" {
				slugTitle = strings.TrimSuffix(base, ".md")
			}
			p.OutputDir = filepath.Join("docs", dir, slugTitle)
		}
	}

	for _, p := range cache.Pages {
		p.MarkdownContent = fixLinksAndImages(cache, p)
	}

	for _, p := range cache.Pages {
		mdParser := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs)
		htmlBytes := markdown.ToHTML(p.MarkdownContent, mdParser, nil)
		p.HTMLContent = template.HTML(htmlBytes)

		if err := renderHTMLPage(cache, p); err != nil {
			return err
		}
	}
	return nil
}

func renderHTMLPage(cache *BuildCache, page *PageData) error {
	if err := os.MkdirAll(page.OutputDir, 0755); err != nil {
		return err
	}
	outFile := filepath.Join(page.OutputDir, "index.html")
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

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
		Page:        page,
		MenuItems:   menuItems,
		MenuTargets: menuTargets,
	}

	tmpl := template.New("page")
	tmpl = initTemplateFuncs(tmpl)
	tmpl, err = tmpl.Parse(htmlTemplate)
	if err != nil {
		return err
	}

	// Debug: Print BasePath to confirm it's loaded
	fmt.Printf("DEBUG: BasePath in renderHTMLPage for page %s: [%s]\n", page.RelPath, cache.Config.Website.BasePath)

	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	fmt.Printf("Generated: %s\n", outFile)
	return nil
}

// generate RSS => docs/rss.xml
func generateRSS(cache *BuildCache) error {
	var dated []*PageData
	for _, p := range cache.Pages {
		if !p.FrontMatter.ParsedDate.IsZero() {
			dated = append(dated, p)
		}
	}
	sort.Slice(dated, func(i, j int) bool {
		return dated[i].FrontMatter.ParsedDate.After(dated[j].FrontMatter.ParsedDate)
	})

	var items []string
	for _, p := range dated {
		pubDate := p.FrontMatter.ParsedDate.Format(time.RFC1123Z)
		title := escapeForXML(p.FrontMatter.Title)
		desc := escapeForXML(p.FrontMatter.Description)

		out := strings.TrimPrefix(p.OutputDir, "docs/")
		link := strings.TrimSuffix(cache.Config.Website.URL, "/") + "/" + out + "/"

		var enclosure string
		if p.FrontMatter.Image != "" {
			img := strings.TrimPrefix(p.FrontMatter.Image, "/")
			enclosure = fmt.Sprintf(`<enclosure url="%s/%s" type="image/png"/>`,
				strings.TrimSuffix(cache.Config.Website.URL, "/"), img)
		}
		items = append(items, fmt.Sprintf(`
<item>
  <title>%s</title>
  <link>%s</link>
  <description>%s</description>
  <pubDate>%s</pubDate>
  %s
</item>
`, title, link, desc, pubDate, enclosure))
	}

	rss := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
  <title>%s</title>
  <link>%s</link>
  <description>RSS feed for %s</description>
  %s
</channel>
</rss>
`, escapeForXML(cache.Config.Website.Name),
		cache.Config.Website.URL,
		escapeForXML(cache.Config.Website.Name),
		strings.Join(items, "\n"))

	if err := os.WriteFile(filepath.Join("docs", "rss.xml"), []byte(rss), 0644); err != nil {
		return err
	}
	fmt.Println("Generated: docs/rss.xml")
	return nil
}

// create 404.html => docs/404.html
func create404Page(cache *BuildCache) error {
	_ = os.MkdirAll("docs", 0755)
	f, err := os.Create(filepath.Join("docs", "404.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	pseudo := &PageData{
		FrontMatter: PageFrontMatter{
			Title: "404 Not Found",
		},
		// HTMLContent will be set after BasePath logic
		RelPath:   "404.html",
		OutputDir: "docs",
	}

	var homeLinkFor404Page string
	if cache.Config.Website.BasePath == "" {
		homeLinkFor404Page = "/"
	} else {
		// Ensure BasePath like "/krems" becomes "/krems/" for the link
		homeLinkFor404Page = strings.TrimSuffix(cache.Config.Website.BasePath, "/") + "/"
	}
	pseudo.HTMLContent = template.HTML(fmt.Sprintf(`
<p>Go <a href="%s">home</a> to find what you're looking for</p>
`, homeLinkFor404Page))

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

	tmpl := template.New("404")
	tmpl = initTemplateFuncs(tmpl)
	tmpl, err = tmpl.Parse(htmlTemplate)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	fmt.Println("Generated: docs/404.html")
	return nil
}

func extractDomain(urlStr string) string {
	parsed, err := url.Parse(strings.TrimSpace(urlStr))
	if err != nil {
		return ""
	}
	// parsed.Host might be "mreider.com" or "www.mreider.com:443"
	// so we trim any port if present
	host := parsed.Host
	if colonPos := strings.Index(host, ":"); colonPos != -1 {
		host = host[:colonPos]
	}
	// if the user typed something like "mreider.com" without scheme,
	// net/url might treat it differently, so we fallback:
	if host == "" && parsed.Path != "" && !strings.Contains(parsed.Path, "/") {
		host = parsed.Path
	}
	return host
}

func escapeForXML(s string) string {
	r := strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&quot;",
		`'`, "&apos;",
	)
	return r.Replace(s)
}

// template funcs
func initTemplateFuncs(t *template.Template) *template.Template {
	return t.Funcs(template.FuncMap{
		"trimPrefixSlash":      trimPrefixSlash,
		"relativeToRoot":       relativeToRoot,
		"imagePath":            imagePath,
		"listPagesInDirectory": listPagesInDirectory,
		"authorLink":           authorLink,
		"tagsLine":             tagsLine,
		"authorLine":           authorLine,
		"dateDisplay":          dateDisplay,
		"sitePath":             sitePath, // Added sitePath
	})
}

func trimPrefixSlash(s string) string {
	return strings.TrimPrefix(s, "/")
}

func relativeToRoot(outputDir string) string {
	rel, _ := filepath.Rel(outputDir, "docs")
	if rel == "." {
		return "."
	}
	return filepath.ToSlash(rel)
}

func imagePath(outputDir, img string) string {
	root := relativeToRoot(outputDir)
	return filepath.ToSlash(filepath.Join(root, img))
}
