package main

import (
	"fmt"
	"html/template"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gosimple/slug"
)

// listPagesInDirectory => for type:list pages at any level
func listPagesInDirectory(relPath string) template.HTML {
	if globalBuildCache == nil {
		return ""
	}

	// 1. Find the listing page
	var listingPage *PageData
	for _, p := range globalBuildCache.Pages {
		if p.RelPath == relPath {
			listingPage = p
			break
		}
	}
	if listingPage == nil {
		// try to find index.md in the directory
		dir := filepath.Dir(relPath)
		if !strings.HasSuffix(dir, "index.md") {
			dir = filepath.Join(dir, "index.md")
		}
		for _, p := range globalBuildCache.Pages {
			if p.RelPath == dir {
				listingPage = p
				break
			}
		}
	}
	if listingPage == nil {
		return ""
	}

	// 2. Determine the directory for that page
	dir := filepath.Dir(listingPage.RelPath)
	if dir == "." {
		dir = ""
	}
	
	// 3. Gather pages with a valid date, skipping index pages
	var siblings []*PageData
	for _, p := range globalBuildCache.Pages {
		// Skip index pages and pages without dates
		if p.IsIndex || p.FrontMatter.ParsedDate.IsZero() {
			continue
		}
		
		include := true
		
		// For author pages, we need to include all posts by the author regardless of directory
		if len(listingPage.FrontMatter.AuthorFilter) > 0 {
			include = false // Default to exclude
			for _, filterAuthor := range listingPage.FrontMatter.AuthorFilter {
				// Case-insensitive comparison for author names
				if strings.EqualFold(strings.TrimSpace(p.FrontMatter.Author), strings.TrimSpace(filterAuthor)) {
					include = true
					break
				}
			}
		} else if len(listingPage.FrontMatter.TagFilter) > 0 {
			// Handle tag filtering
			include = false // Default to exclude
			for _, pageTag := range p.FrontMatter.Tags {
				for _, filterTag := range listingPage.FrontMatter.TagFilter {
					// Case-insensitive comparison for tags
					if strings.EqualFold(strings.TrimSpace(pageTag), strings.TrimSpace(filterTag)) {
						include = true
						break
					}
				}
				if include {
					break
				}
			}
		} else {
			// For regular list pages, only include pages from the same directory
			pageDir := filepath.Dir(p.RelPath)
			if pageDir == "." {
				pageDir = ""
			}
			
			// Check if page is in the same directory as the listing page
			if pageDir != dir {
				include = false
			}
		}

		if include {
			siblings = append(siblings, p)
		}
	}
	
	// 4. Sort descending by date
	sort.Slice(siblings, func(i, j int) bool {
		return siblings[i].FrontMatter.ParsedDate.After(siblings[j].FrontMatter.ParsedDate)
	})

	// 5. Group by year=>month, then build HTML
	groups := groupByYearThenMonth(siblings)

	var sb strings.Builder
	sb.WriteString(`<div class="blog-list">`)
	for _, yg := range groups {
		sb.WriteString(fmt.Sprintf(`<h3 class="mt-5 mb-3">%d</h3>`+"\n", yg.Year))
		for _, mg := range yg.Months {
			sb.WriteString(fmt.Sprintf(`<h5 class="mb-2">%s</h5>`+"\n", mg.Month))
			sb.WriteString(`<ul class="list-group mb-4" style="padding-left: 20px; margin-left: 0;">` + "\n")
			for _, art := range mg.Pages {
				outDir := FindPageByRelPath(globalBuildCache, art.RelPath)
				// Construct the path part first, then pass to sitePath
				pageLinkPath := "/" + outDir + "/"
				finalPageLink := sitePath(pageLinkPath)
				
				authorText := ""
				if art.FrontMatter.Author != "" {
					authorSlug := slug.Make(art.FrontMatter.Author)
					authorLinkPath := "/authors/" + authorSlug + "/"
					authorText = fmt.Sprintf(` by <a href="%s">%s</a>`, sitePath(authorLinkPath), art.FrontMatter.Author)
				}
				
				tagsText := ""
				if len(art.FrontMatter.Tags) > 0 {
					var tagLinks []string
					for _, tag := range art.FrontMatter.Tags {
						tagSlug := slug.Make(tag)
						tagLinkPath := "/tags/" + tagSlug + "/"
						tagLinks = append(tagLinks, fmt.Sprintf(`<a href="%s" class="tag-link"><span class="tag-badge">%s</span></a>`, sitePath(tagLinkPath), tag))
					}
					tagsText = strings.Join(tagLinks, " ")
				}
				sb.WriteString(fmt.Sprintf(
					`<li><a class="text-decoration-none" href="%s">%s</a> <span class="text-muted small">%s %s</span></li>`+"\n",
					finalPageLink, art.FrontMatter.Title, authorText, tagsText))
			}
			sb.WriteString("</ul>\n")
		}
	}
	sb.WriteString(`</div>`)
	return template.HTML(sb.String())
}

// groupByYearThenMonth lumps pages by Year => Month => Pages
func groupByYearThenMonth(pages []*PageData) []yearGroup {
	yMap := make(map[int]map[time.Month][]*PageData)
	for _, p := range pages {
		y := p.FrontMatter.ParsedDate.Year()
		m := p.FrontMatter.ParsedDate.Month()
		if _, ok := yMap[y]; !ok {
			yMap[y] = make(map[time.Month][]*PageData)
		}
		yMap[y][m] = append(yMap[y][m], p)
	}

	// sort years descending
	var years []int
	for year := range yMap {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))

	var result []yearGroup
	for _, yr := range years {
		monthMap := yMap[yr]
		var months []time.Month
		for mm := range monthMap {
			months = append(months, mm)
		}
		sort.Slice(months, func(i, j int) bool {
			return months[i] > months[j]
		})

		var mgs []monthGroup
		for _, mm := range months {
			mgs = append(mgs, monthGroup{
				Month: mm,
				Pages: monthMap[mm],
			})
		}
		result = append(result, yearGroup{
			Year:   yr,
			Months: mgs,
		})
	}
	return result
}

// yearGroup => one year => multiple months
type yearGroup struct {
	Year   int
	Months []monthGroup
}

// monthGroup => one month => pages
type monthGroup struct {
	Month time.Month
	Pages []*PageData
}
