package main

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/gosimple/slug"
)

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{if .Page.FrontMatter.Title}}{{.Page.FrontMatter.Title}} - {{end}}{{.Config.Website.Name}}</title>
    <meta name="description" content="{{.Page.FrontMatter.Description}}">
    <meta name="viewport" content="width:device-width, initial-scale=1.0">
    {{if .Page.FrontMatter.Image}}
    <meta property="og:image" content="{{.Config.Website.URL}}{{sitePath (.Page.FrontMatter.Image | trimPrefixSlash)}}" />
    <meta property="og:image:width" content="1200" />
    <meta property="og:image:height" content="630" />
    
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:image" content="{{.Config.Website.URL}}{{sitePath (.Page.FrontMatter.Image | trimPrefixSlash)}}" />
    {{end}}
    <meta property="og:site_name" content="{{.Config.Website.Name}}">
    <link rel="icon" href="{{sitePath "/images/favicon.ico"}}" type="image/x-icon">

    {{if .AlternativeCSSFiles}}
        {{range .AlternativeCSSFiles}}
    <link rel="stylesheet" href="{{sitePath .}}">
        {{end}}
    {{else}}
    <link rel="stylesheet" href="{{sitePath "/css/bootstrap.min.css"}}">
    <link rel="stylesheet" href="{{sitePath "/css/custom.css"}}">
    {{end}}
</head>
<body>

<!-- Wrap the navbar in the container for consistency -->
<nav class="navbar navbar-expand-lg navbar-light">
    <div class="container"> <!-- Added container class here for consistent alignment -->
        <!-- Burger Menu (Still Functional) -->
        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#kremsNavbar" 
                aria-controls="kremsNavbar" aria-expanded="false" aria-label="Toggle navigation">
            <span class="navbar-toggler-icon"></span>
        </button>

        <!-- Navbar Links on the Left -->
        <div class="collapse navbar-collapse" id="kremsNavbar">
            <ul class="navbar-nav me-auto mb-2 mb-lg-0">
                {{range $i, $targetPath := .MenuTargets}}
                <li class="nav-item">
                    <a class="nav-link" href="{{sitePath $targetPath}}">{{index $.MenuItems $i}}</a>
                </li>
                {{end}}
            </ul>
        </div>

        <!-- Website Title on the Right -->
        <a class="navbar-brand" href="{{sitePath "/"}}">
            {{.Config.Website.Name}}
        </a>
    </div>
</nav>

<!-- Content -->
<div class="container-lg mt-5 mb-5">
    {{if .Page.FrontMatter.Image}}
    {{ $cleanImg := .Page.FrontMatter.Image | trimPrefixSlash }}
    <img src="{{sitePath $cleanImg}}" style="max-width:400px;width:100%;height:auto;" class="img-fluid mb-3 rounded" alt="featured image">
    {{end}}

	{{if (ne .Page.FrontMatter.Title "")}}
    <h3 class="display-6 mb-4">{{.Page.FrontMatter.Title}}</h3>
	{{authorLine .Page.FrontMatter.Author}}
	{{dateDisplay .Page.FrontMatter.ParsedDate}}
	{{tagsLine .Page.FrontMatter.Tags}}
    {{end}}
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

    <footer class="text-center mt-5">
        Generated with <a href="https://github.com/mreider/krems">Krems</a>
    </footer>
</div>

{{if .AlternativeJSFiles}}
    {{range .AlternativeJSFiles}}
<script src="{{sitePath .}}"></script>
    {{end}}
{{else}}
<script src="{{sitePath "/js/bootstrap.js"}}"></script>
{{end}}
</body>
</html>
`

// authorLink generates a link to the author's page
func authorLink(author string) template.HTML {
	if author == "" {
		return ""
	}
	authorSlug := slug.Make(author)
	// Note: sitePath expects paths like "/authors/..."
	return template.HTML(fmt.Sprintf(` by <a href="%s">%s</a>`, sitePath("/authors/"+authorSlug+"/"), author))
}

// tagsLine generates a list of tags with links to tag pages
func tagsLine(tags []string) template.HTML {
	if len(tags) == 0 {
		return ""
	}
	var tagLinks []string
	for _, tag := range tags {
		tagSlug := slug.Make(tag)
		// Note: sitePath expects paths like "/tags/..."
		tagLinks = append(tagLinks, fmt.Sprintf(`<a href="%s" class="tag-link"><span class="tag-badge">%s</span></a>`, sitePath("/tags/"+tagSlug+"/"), tag))
	}
	return template.HTML(strings.Join(tagLinks, " "))
}

// authorLine generates the author line with a link to the author's page
func authorLine(author string) template.HTML {
	if author == "" {
		return ""
	}
	authorSlug := slug.Make(author)
	// Note: sitePath expects paths like "/authors/..."
	return template.HTML(fmt.Sprintf(`by <a href="%s">%s</a>`, sitePath("/authors/"+authorSlug+"/"), author))
}

// dateDisplay formats the date in a nice format (Jan 1, 2025)
func dateDisplay(date time.Time) template.HTML {
	if date.IsZero() {
		return ""
	}
	return template.HTML(fmt.Sprintf(`<div class="text-muted mb-2">%s</div>`, date.Format("Jan 2, 2006")))
}

// sitePath prepends the BasePath to a given path, ensuring correct slash handling.
// path argument should typically start with a slash (e.g., "/css/style.css", "/my-page/").
func sitePath(path string) string {
	if globalBuildCache == nil || globalBuildCache.Config == nil {
		// Fallback if cache not ready, though should not happen in normal execution
		return path 
	}
	basePath := globalBuildCache.Config.Website.BasePath // e.g., "/krems" or ""

	// Ensure basePath is clean (no trailing slash if not root, unless it IS the root path "/")
	// Ensure path starts with a slash if not empty
	// This logic aims for:
	// basePath="/krems", path="/css/style.css" -> "/krems/css/style.css"
	// basePath="", path="/css/style.css" -> "/css/style.css"
	// basePath="/krems", path="/" -> "/krems/" (for homepage)
	// basePath="", path="/" -> "/" (for homepage)

	// Normalize path to always start with a slash if it's not empty
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	// If basePath is empty or just "/", effectively no prefix needed or path is already root-relative
	if basePath == "" {
		return path
	}

	// If basePath is like "/krems"
	// Remove trailing slash from basePath if it exists and isn't the only char
	cleanBasePath := strings.TrimSuffix(basePath, "/")
	
	// If path is just "/", append to cleaned basePath (e.g. /krems + / -> /krems/)
	if path == "/" {
		if cleanBasePath == "" { // Original basePath was "/"
			return "/"
		}
		return cleanBasePath + "/"
	}

	// Standard join: /krems + /css/style.css -> /krems/css/style.css
	return cleanBasePath + path
}
