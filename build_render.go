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

	homePathFor404 := cache.Config.Website.BasePath + "/"
	if cache.Config.Website.BasePath == "" {
		// Ensure it's just a single slash if BasePath is empty
		homePathFor404 = "/"
	} else if !strings.HasSuffix(cache.Config.Website.BasePath, "/") {
		// Ensure BasePath (if not empty) ends with a slash for consistency with other links
		// This might be redundant if BasePath is always stored like "/krems" or ""
		// but good for safety. However, other links add it: {{BasePath}}/slug/
		// So, if BasePath is /krems, then /krems/ is correct for home.
		// If BasePath is "", then / is correct.
		// The template uses {{.Config.Website.BasePath}}/, so let's match that.
		// If BasePath is /foo, then /foo/. If BasePath is "", then /.
		// This is already handled by `href="{{.Config.Website.BasePath}}/"` in the main template.
		// So, for consistency:
		// homePathFor404 = cache.Config.Website.BasePath 
		// if homePathFor404 != "" && !strings.HasSuffix(homePathFor404, "/") {
		// 	homePathFor404 += "/"
		// } else if homePathFor404 == "" {
		//  homePathFor404 = "/"
		// }
		// This logic is simpler:
		// if BasePath is "/foo", then "/foo/"
		// if BasePath is "", then "/"
	}
	// The navbar brand link uses {{.Config.Website.BasePath}}/, so let's be consistent.
	// If BasePath is "/foo", this becomes "/foo/"
	// If BasePath is "", this becomes "/"
	// This logic for homePathFor404 was a bit convoluted and not strictly necessary
	// as the fmt.Sprintf below handles it well with cache.Config.Website.BasePath + "/"
	// (empty string + "/" is still "/", and "/foo" + "/" is "/foo/")
	// However, the direct use of BasePath in Sprintf is fine.
	// The critical fix is removing the misplaced '}'
	pseudo.HTMLContent = template.HTML(fmt.Sprintf(`
<p>Go <a href="%s/">home</a> to find what you're looking for</p>
`, cache.Config.Website.BasePath))
// Removed erroneous closing brace that was here.

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
