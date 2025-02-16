package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gosimple/slug"
	"gopkg.in/yaml.v3"
)

// Config matches the structure of config.yaml

type QuackerConfig struct {
	Domain    string `yaml:"domain"`
	SiteOwner string `yaml:"site_owner"`
	Target    string `yaml:"target"`
}

type Config struct {
	Website struct {
		URL  string `yaml:"url"`
		Name string `yaml:"name"`
	} `yaml:"website"`
	Menu []struct {
		Title string `yaml:"title"`
		Path  string `yaml:"path"`
	} `yaml:"menu"`

	Quacker *QuackerConfig `yaml:"quacker,omitempty"`
}

// PageFrontMatter is the front matter in each .md file
type PageFrontMatter struct {
	Title       string `yaml:"title"`
	Type        string `yaml:"type"` // normal|list
	Description string `yaml:"description"`
	Image       string `yaml:"image"` // "/images/foo.png" or "images/foo.png"
	Date        string `yaml:"date"`
	ParsedDate  time.Time
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

// handleBuild => krems --build
func handleBuild() {
	// remove docs/ if exists
	_ = os.RemoveAll("docs")

	// read config.yaml
	cfg, err := readConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error reading config.yaml: %v\n", err)
		os.Exit(1)
	}

	// copy static assets from markdown/(css|js|images) => docs/
	if err := copyStaticAssets(); err != nil {
		fmt.Printf("Error copying static assets: %v\n", err)
		os.Exit(1)
	}

	// parse all .md => PageData
	pages, err := parseMarkdownFiles("markdown")
	if err != nil {
		fmt.Printf("Error parsing markdown: %v\n", err)
		os.Exit(1)
	}

	// create BuildCache
	cache := &BuildCache{
		Pages:  pages,
		Config: cfg,
	}
	assignGlobalCache(cache)

	// process pages => rewrite links => HTML => final
	if err := processPages(cache); err != nil {
		fmt.Printf("Error processing pages: %v\n", err)
		os.Exit(1)
	}

	domain := extractDomain(cache.Config.Website.URL)
	if domain != "" {
		cnameFile := filepath.Join("docs", "CNAME")
		err := os.WriteFile(cnameFile, []byte(domain+"\n"), 0644)
		if err != nil {
			fmt.Printf("Error creating CNAME file: %v\n", err)
			// Decide if you want to exit or just print the error
		} else {
			fmt.Printf("Created: %s (CNAME)\n", cnameFile)
		}
	}

	// generate rss.xml
	if err := generateRSS(cache); err != nil {
		fmt.Printf("Error generating RSS: %v\n", err)
		os.Exit(1)
	}

	// create 404.html
	if err := create404Page(cache); err != nil {
		fmt.Printf("Error creating 404.html: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Build complete! The 'docs/' directory is ready.")
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

func readConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func copyStaticAssets() error {
	subdirs := []string{"css", "js", "images"}
	for _, sd := range subdirs {
		src := filepath.Join("markdown", sd)
		dest := filepath.Join("docs", sd)
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

// process => compute output dirs, fix .md => /slug, 400px images => HTML
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
				`<img src="%s" alt="%s" style="max-width:400px;width:100%%;height:auto;" class="mb-3 img-fluid"/>`,
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
				// if no slash, assume same dir => join with page's dir
				linkCandidate := linkTarget
				if !strings.Contains(linkCandidate, "/") {
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
		HTMLContent: template.HTML(`
<p>Go <a href="/">home</a> to find what you're looking for</p>
`),
		RelPath:   "404.html",
		OutputDir: "docs",
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

// The main HTML template references bootstrap absolutely => "/css/bootstrap.min.css"
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>{{if .Page.FrontMatter.Title}}{{.Page.FrontMatter.Title}} - {{end}}{{.Config.Website.Name}}</title>
	<meta name="description" content="{{.Page.FrontMatter.Description}}">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	{{if .Page.FrontMatter.Image}}
	<meta property="og:image" content="{{.Page.FrontMatter.Image | trimPrefixSlash}}">
	<meta name="twitter:image" content="{{.Page.FrontMatter.Image | trimPrefixSlash}}">
	{{end}}
	<meta property="og:site_name" content="{{.Config.Website.Name}}">
	<link rel="icon" href="/images/favicon.png" type="image/png">

	<link rel="stylesheet" href="/css/bootstrap.min.css">
	<style>
	body{background: #FBF9F1;}
	</style>
</head>
<body>
<nav class="navbar navbar-expand-lg navbar-light" style="background-color:#E5E1DA; margin-bottom:20px;">
	<div class="container-fluid">
		<a class="navbar-brand" href="/">
			<img src="/images/favicon.png" alt="" width="30" height="30" class="d-inline-block align-text-top me-1">
			{{.Config.Website.Name}}
		</a>
		<button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#kremsNavbar" aria-controls="kremsNavbar" aria-expanded="false" aria-label="Toggle navigation">
			<span class="navbar-toggler-icon"></span>
		</button>
		<div class="collapse navbar-collapse" id="kremsNavbar">
			<ul class="navbar-nav ms-auto mb-2 mb-lg-0">
				{{range $i, $label := .MenuItems}}
				<li class="nav-item">
					<a class="nav-link" href="{{index $.MenuTargets $i}}">{{$label}}</a>
				</li>
				{{end}}
			</ul>
		{{ if and .Config.Quacker (ne .Config.Quacker.Target "") }}
		<form id="subscribe-form" action="https://{{.Config.Quacker.Target}}/subscribe" method="POST" class="d-flex ms-3">
			<input type="hidden" name="owner" value="{{.Config.Quacker.SiteOwner}}">
			<input type="hidden" name="domain" value="{{.Config.Quacker.Domain}}">
			<div class="d-flex align-items-center gap-2" id="form-content">
				<input type="email" class="form-control form-control-sm" id="email" name="email" placeholder="email" required>
				<button type="submit" class="btn btn-secondary btn-sm">Subscribe</button>
			</div>
		</form>
		<div id="subscribe-message" class="mt-3"></div>

		<script>
		document.getElementById('subscribe-form').addEventListener('submit', function(event) {
			event.preventDefault();
			
			var form = document.getElementById('subscribe-form');
			var messageDiv = document.getElementById('subscribe-message');

			// Hide the form using Bootstrap's d-none class
			form.classList.add('d-none');

			// Clear previous messages
			messageDiv.innerHTML = '';

			var formData = new FormData(form);
			fetch(form.action, {
				method: 'POST',
				body: formData
			})
			.then(response => {
				if (response.ok) {
					return response.text();
				}
				throw new Error('Subscription failed');
			})
			.then(message => {
				messageDiv.innerHTML = '<div class="alert alert-success">' + message + '</div>';
				setTimeout(() => {
					messageDiv.innerHTML = '';  // Clear message
					form.classList.remove('d-none'); // Show form again
					form.reset();
				}, 3000);
			})
			.catch(error => {
				messageDiv.innerHTML = '<div class="alert alert-danger">' + error.message + '</div>';
				setTimeout(() => {
					messageDiv.innerHTML = '';  // Clear message
					form.classList.remove('d-none'); // Show form again
				}, 3000);
			});
		});
		</script>

		{{ end }}
		</div>
	</div>
</nav>

<div class="container mt-5 mb-5">
{{if .Page.FrontMatter.Image}}
	{{ $cleanImg := .Page.FrontMatter.Image | trimPrefixSlash }}
	<img src="/{{$cleanImg}}" style="max-width:400px;width:100%;height:auto;" class="img-fluid mb-3 rounded" alt="featured image">
{{end}}

<h3 class="display-6 mb-4">{{if .Page.FrontMatter.Title}}{{.Page.FrontMatter.Title}}{{else}}{{.Config.Website.Name}}{{end}}</h3>
{{if .Page.FrontMatter.Description}}
<p class="text-muted mb-4">{{.Page.FrontMatter.Description}}</p>
{{end}}

{{if eq .Page.FrontMatter.Type "list"}}
	{{listPagesInDirectory .Page.RelPath}}
{{else}}
	<div class="mb-5">
		{{.Page.HTMLContent}}
	</div>
{{end}}

<footer class="text-center">
	♜ Generated with <a href="https://github.com/mreider/krems">Krems</a> ♜
</footer>
</div>

<script src="/js/bootstrap.js"></script>
</body>
</html>
`
