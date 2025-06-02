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
    <meta property="og:image" content="{{.Config.Website.URL}}{{.Config.Website.BasePath}}/{{.Page.FrontMatter.Image | trimPrefixSlash}}" />
    <meta property="og:image:width" content="1200" />
    <meta property="og:image:height" content="630" />
    
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:image" content="{{.Config.Website.URL}}{{.Config.Website.BasePath}}/{{.Page.FrontMatter.Image | trimPrefixSlash}}" />
    {{end}}
    <meta property="og:site_name" content="{{.Config.Website.Name}}">
    <link rel="icon" href="{{.Config.Website.BasePath}}/images/favicon.ico" type="image/x-icon">

    <link rel="stylesheet" href="{{.Config.Website.BasePath}}/css/bootstrap.min.css">
    <style>
        /* Load custom font */
        @font-face {
            font-family: 'Tiempos';
            src: url('{{.Config.Website.BasePath}}/css/tiempos.woff2') format('woff2');
            font-weight: normal;
            font-style: normal;
        }

        /* Apply the custom font globally */
        :root {
            --bs-body-font-family: 'Tiempos', serif;
        }

        /* Remove all background colors */
        body, nav {
            background-color: white !important;
        }

        /* Adjust font sizes subtly */
        body {
            font-size: 1.1rem;
        }

        h3 {
            font-size: 1.8rem;
        }

        /* Navbar adjustments */
        .navbar-brand {
            font-weight: bold;
        }

        /* Ensure the navbar is aligned with content */
        .navbar-nav {
            width: auto;
        }

        .navbar {
            justify-content: space-between;
            width: 100%;
        }

        /* Align navbar with content using the same container width */
        .navbar .container {
            max-width: 960px;
            margin: 0 auto;
            padding: 0 15px; /* Adds a bit of padding to prevent edge hugging */
        }

        /* Fix mobile menu layout to display vertically when collapsed */
        @media (max-width: 991.98px) {
            .navbar-collapse .navbar-nav {
                flex-direction: column !important;
                align-items: flex-start !important;
                margin-left: 0 !important;
            }
            
            .navbar-collapse .navbar-nav .nav-item {
                margin-bottom: 10px;
                width: 100%;
            }
            
            .navbar-collapse .nav-link {
                padding: 8px 0;
            }
        }

        /* Desktop menu layout */
        @media (min-width: 992px) {
            .navbar .navbar-nav {
                flex-direction: row;
                align-items: center;
            }
            
            .navbar-collapse .navbar-nav {
                margin-left: 10px;
            }
            
            .navbar .navbar-nav .nav-item {
                margin-right: 15px;
            }
        }

        /* Fix for bullets in content area, ensure they are inside */
        ul {
            padding-left: 20px; /* Indentation for the bullets */
            margin-left: 0; /* Remove any negative margin */
        }

        /* Centered and well-spaced Quacker form */
        .subscribe-container {
            text-align: center;
            margin-bottom: 40px;
        }

        /* Fix the left align of content and navbar */
        .container-lg {
            max-width: 960px;
            margin-left: auto;
            margin-right: auto;
            padding-left: 15px;
            padding-right: 15px;
        }
    </style>
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
                {{range $i, $label := .MenuItems}}
                <li class="nav-item">
                    <a class="nav-link" href="{{$.Config.Website.BasePath}}{{index $.MenuTargets $i}}">{{$label}}</a>
                </li>
                {{end}}
            </ul>
        </div>

        <!-- Website Title on the Right -->
        <a class="navbar-brand" href="{{.Config.Website.BasePath}}/">
            {{.Config.Website.Name}}
        </a>
    </div>
</nav>

<!-- Content -->
<div class="container-lg mt-5 mb-5">
    {{if .Page.FrontMatter.Image}}
    {{ $cleanImg := .Page.FrontMatter.Image | trimPrefixSlash }}
    <img src="{{$.Config.Website.BasePath}}/{{$cleanImg}}" style="max-width:400px;width:100%;height:auto;" class="img-fluid mb-3 rounded" alt="featured image">
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

    {{ if and .Config.Quacker (ne .Config.Quacker.Target "") }}
    <div class="subscribe-container">
        <form id="subscribe-form" action="https://{{.Config.Quacker.Target}}/subscribe" method="POST">
            <input type="hidden" name="owner" value="{{.Config.Quacker.SiteOwner}}">
            <input type="hidden" name="domain" value="{{.Config.Quacker.Domain}}">
            <div class="d-flex justify-content-center align-items-center gap-2">
                <input type="email" class="form-control form-control-sm w-auto" id="email" name="email" placeholder="email" required>
                <button type="submit" class="btn btn-secondary btn-sm">Subscribe</button>
            </div>
        </form>
        <div id="subscribe-message" class="mt-3"></div>
    </div>

    <script>
    document.getElementById('subscribe-form').addEventListener('submit', function(event) {
        event.preventDefault();
        
        var form = document.getElementById('subscribe-form');
        var messageDiv = document.getElementById('subscribe-message');

        form.classList.add('d-none');
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
                form.classList.remove('d-none');
                form.reset();
            }, 3000);
        })
        .catch(error => {
            messageDiv.innerHTML = '<div class="alert alert-danger">' + error.message + '</div>';
            setTimeout(() => {
                messageDiv.innerHTML = '';  // Clear message
                form.classList.remove('d-none');
            }, 3000);
        });
    });
    </script>
    {{ end }}

    <footer class="text-center mt-5">
        Generated with <a href="https://github.com/mreider/krems">Krems</a>
    </footer>
</div>

<script src="{{.Config.Website.BasePath}}/js/bootstrap.js"></script>
</body>
</html>
`

// authorLink generates a link to the author's page
func authorLink(author string) template.HTML {
	if author == "" || globalBuildCache == nil || globalBuildCache.Config == nil {
		return ""
	}
	authorSlug := slug.Make(author)
	basePath := globalBuildCache.Config.Website.BasePath
	// Ensure no double slashes if basePath is empty, but always a leading slash if basePath is not empty.
	// And ensure a trailing slash for the directory.
	return template.HTML(fmt.Sprintf(` by <a href="%s/authors/%s/">%s</a>`, basePath, authorSlug, author))
}

// tagsLine generates a list of tags with links to tag pages
func tagsLine(tags []string) template.HTML {
	if len(tags) == 0 || globalBuildCache == nil || globalBuildCache.Config == nil {
		return ""
	}
	basePath := globalBuildCache.Config.Website.BasePath
	var tagLinks []string
	for _, tag := range tags {
		tagSlug := slug.Make(tag)
		tagLinks = append(tagLinks, fmt.Sprintf(`<a href="%s/tags/%s/"><span class="badge bg-secondary">%s</span></a>`, basePath, tagSlug, tag))
	}
	return template.HTML(strings.Join(tagLinks, " "))
}

// authorLine generates the author line with a link to the author's page
func authorLine(author string) template.HTML {
	if author == "" || globalBuildCache == nil || globalBuildCache.Config == nil {
		return ""
	}
	basePath := globalBuildCache.Config.Website.BasePath
	authorSlug := slug.Make(author)
	return template.HTML(fmt.Sprintf(`by <a href="%s/authors/%s/">%s</a>`, basePath, authorSlug, author))
}

// dateDisplay formats the date in a nice format (Jan 1, 2025)
func dateDisplay(date time.Time) template.HTML {
	if date.IsZero() {
		return ""
	}
	return template.HTML(fmt.Sprintf(`<div class="text-muted mb-2">%s</div>`, date.Format("Jan 2, 2006")))
}
