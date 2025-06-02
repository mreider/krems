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

func processPages(cache *BuildCache, outputDirRoot string) error { // MODIFIED: Added outputDirRoot
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
			p.OutputDir = filepath.Join(outputDirRoot, dir) // MODIFIED: Used outputDirRoot
		} else {
			slugTitle := slug.Make(p.FrontMatter.Title)
			if slugTitle == "" {
				slugTitle = strings.TrimSuffix(base, ".md")
			}
			p.OutputDir = filepath.Join(outputDirRoot, dir, slugTitle) // MODIFIED: Used outputDirRoot
		}
	}

	for _, p := range cache.Pages {
		p.MarkdownContent = fixLinksAndImages(cache, p)
	}

	for _, p := range cache.Pages {
		mdParser := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs)
		htmlBytes := markdown.ToHTML(p.MarkdownContent, mdParser, nil)
		p.HTMLContent = template.HTML(htmlBytes)

		// MODIFIED: Pass outputDirRoot (as siteBuildRoot) to renderHTMLPage
		if err := renderHTMLPage(cache, p, outputDirRoot); err != nil {
			return err
		}
	}
	return nil
}

// MODIFIED: Added siteBuildRoot parameter
func renderHTMLPage(cache *BuildCache, page *PageData, siteBuildRoot string) error {
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
	// MODIFIED: Pass siteBuildRoot to initTemplateFuncs
	tmpl = initTemplateFuncs(tmpl, siteBuildRoot)
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

// generate RSS => outputDirRoot/rss.xml
func generateRSS(cache *BuildCache, outputDirRoot string) error { // MODIFIED: Added outputDirRoot
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

		// p.OutputDir is now like /tmp/krems-run-XYZ/actual/path
		// We need to make it relative to outputDirRoot for the URL
		relOut, err := filepath.Rel(outputDirRoot, p.OutputDir)
		if err != nil {
			// This should not happen if p.OutputDir is correctly prefixed
			return fmt.Errorf("could not make path %s relative to %s: %w", p.OutputDir, outputDirRoot, err)
		}
		link := strings.TrimSuffix(cache.Config.Website.URL, "/") + "/" + filepath.ToSlash(relOut) + "/"
		if p.IsIndex && relOut == "." { // Handle root index.md case
			link = strings.TrimSuffix(cache.Config.Website.URL, "/") + "/"
		}


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

	rssPath := filepath.Join(outputDirRoot, "rss.xml") // MODIFIED: Used outputDirRoot
	if err := os.WriteFile(rssPath, []byte(rss), 0644); err != nil {
		return err
	}
	fmt.Printf("Generated: %s\n", rssPath) // MODIFIED
	return nil
}

// create 404.html => outputDirRoot/404.html
func create404Page(cache *BuildCache, outputDirRoot string) error { // MODIFIED: Added outputDirRoot
	_ = os.MkdirAll(outputDirRoot, 0755) // MODIFIED: Used outputDirRoot
	f, err := os.Create(filepath.Join(outputDirRoot, "404.html")) // MODIFIED: Used outputDirRoot
	if err != nil {
		return err
	}
	defer f.Close()

	pseudo := &PageData{
		FrontMatter: PageFrontMatter{
			Title: "404 Not Found",
		},
		// HTMLContent will be set after BasePath logic
		RelPath:   "404.html", // This is fine, it's a pseudo path
		OutputDir: outputDirRoot, // MODIFIED: Used outputDirRoot
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

	fmt.Printf("Generated: %s\n", filepath.Join(outputDirRoot, "404.html")) // MODIFIED
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
// MODIFIED: initTemplateFuncs already changed to accept siteBuildRoot in the previous step. This is just for context.
// func initTemplateFuncs(t *template.Template, siteBuildRoot string) *template.Template {
// 	return t.Funcs(template.FuncMap{
// 		"trimPrefixSlash":      trimPrefixSlash,
// 		"relativeToRoot":       func(pageOutputDir string) string { return relativeToRoot(pageOutputDir, siteBuildRoot) },
// 		"imagePath":            func(pageOutputDir, img string) string { return imagePath(pageOutputDir, siteBuildRoot, img) },
// 		"listPagesInDirectory": listPagesInDirectory, // This might also need siteBuildRoot if it constructs paths
// 		"authorLink":           authorLink,
// 		"tagsLine":             tagsLine,
		"authorLine":           authorLine,
		"dateDisplay":          dateDisplay,
		"sitePath":             sitePath, // Added sitePath
	})
}

func trimPrefixSlash(s string) string {
	return strings.TrimPrefix(s, "/")
}

func relativeToRoot(pageOutputDir, siteBuildRoot string) string { // MODIFIED: Added siteBuildRoot
	// Ensure siteBuildRoot is absolute or resolvable correctly with pageOutputDir
	// For simplicity, assume both are absolute or consistently relative for Rel to work.
	// If pageOutputDir is /tmp/build/foo and siteBuildRoot is /tmp/build, rel should be ".."
	// If pageOutputDir is /tmp/build/foo/bar and siteBuildRoot is /tmp/build, rel should be "../.."

	// If pageOutputDir is identical to siteBuildRoot (e.g. for index.html at root)
	if filepath.Clean(pageOutputDir) == filepath.Clean(siteBuildRoot) {
		return "."
	}

	rel, err := filepath.Rel(pageOutputDir, siteBuildRoot)
	if err != nil {
		// Fallback or error handling if paths are not relatable (e.g. different drives on Windows)
		// For now, let's assume they are relatable.
		// This might happen if pageOutputDir is not a subpath of siteBuildRoot, which would be an error.
		// Or if siteBuildRoot is not "above" or at the same level as pageOutputDir.
		// Let's print an error and return a sensible default or panic.
		fmt.Fprintf(os.Stderr, "Error calculating relative path from %s to %s: %v\n", pageOutputDir, siteBuildRoot, err)
		return "." // Fallback, might lead to broken links
	}

	// filepath.Rel might return something like `../..`
	// This should be correct for constructing paths from the page's location back to the root.
	if rel == "." { // This case might be covered by the Clean check above, but good to have.
		return "." // Already at the root or same level
	}
	return filepath.ToSlash(rel)
}

// imagePath calculates the path to an image relative to the current page's output directory.
// pageOutputDir is the directory where the current HTML page is being written (e.g., /tmp/build/posts/my-post).
// siteBuildRoot is the root of the build output (e.g., /tmp/build).
// img is the path to the image relative to the siteBuildRoot (e.g., images/my-image.png or /images/my-image.png).
func imagePath(pageOutputDir, siteBuildRoot, img string) string {
	// img can be /images/foo.png or images/foo.png. We want it relative to siteBuildRoot.
	imgRelToSiteRoot := strings.TrimPrefix(img, "/")

	// Path of the image file on the filesystem
	absImgPath := filepath.Join(siteBuildRoot, imgRelToSiteRoot)

	// Path from the current page's directory to the image file
	relPathToImg, err := filepath.Rel(pageOutputDir, absImgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calculating relative image path from %s to %s: %v\n", pageOutputDir, absImgPath, err)
		return img // Fallback
	}
	return filepath.ToSlash(relPathToImg)
}

// sitePath needs to be aware of the siteBuildRoot to correctly use relativeToRoot
// and also the actual page's output directory.
// The current global cache `globalCache` has `Config.Website.BasePath`.
// The `page` object has `OutputDir`.
// We need to pass `siteBuildRoot` to the template functions.

// Modified initTemplateFuncs to pass siteBuildRoot
func initTemplateFuncs(t *template.Template, siteBuildRoot string) *template.Template {
	return t.Funcs(template.FuncMap{
		"trimPrefixSlash":      trimPrefixSlash,
		"relativeToRoot":       func(pageOutputDir string) string { return relativeToRoot(pageOutputDir, siteBuildRoot) },
		"imagePath":            func(pageOutputDir, img string) string { return imagePath(pageOutputDir, siteBuildRoot, img) },
		"listPagesInDirectory": listPagesInDirectory, // This might also need siteBuildRoot if it constructs paths
		"authorLink":           authorLink,
		"tagsLine":             tagsLine,
		"authorLine":           authorLine,
		"dateDisplay":          dateDisplay,
		"sitePath":             func(p string) string { return sitePath(globalCache.Config.Website.BasePath, p) }, // sitePath itself seems okay if BasePath is correct
	})
}

// renderHTMLPage needs to pass siteBuildRoot to initTemplateFuncs
// The siteBuildRoot is the `page.OutputDir`'s root part, which is `outputDirRoot` from `processPages`.
// However, renderHTMLPage is called for each page, and `page.OutputDir` is specific.
// The `siteBuildRoot` is the one passed to `handleBuild`.
// So, `renderHTMLPage` needs `siteBuildRoot` as a parameter.
// And `processPages` needs to pass it to `renderHTMLPage`.

// Let's adjust renderHTMLPage signature and call
// renderHTMLPage(cache *BuildCache, page *PageData, siteBuildRoot string)

// In processPages:
// if err := renderHTMLPage(cache, p, outputDirRoot); err != nil { return err }

// Then in renderHTMLPage:
// tmpl = initTemplateFuncs(tmpl, siteBuildRoot)
// This seems more robust.

// The definition of sitePath is not in this file, I'll assume it's in htmlTemplate.go or similar
// and correctly uses BasePath. The main concern is that BasePath itself is set correctly
// (which handleBuild does: `if isDevMode && cfg.Website.DevPath != "" { cfg.Website.BasePath = cfg.Website.DevPath }`)
// and that relative paths generated by `imagePath` etc. work from the served location.
