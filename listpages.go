package main

import (
	"fmt"
	"html/template"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
		return ""
	}

	// 2. Determine the directory for that page
	dir := filepath.Dir(listingPage.RelPath)
	if dir == "." {
		dir = ""
	}
	// fmt.Printf("[DEBUG] listingPage: %s, dir: %q\n", relPath, dir)

	// 3. Gather siblings with a valid date, skipping index pages
	var siblings []*PageData
	for _, p := range globalBuildCache.Pages {
		// Also handle top-level: e.g. filepath.Dir("krems_city_info.md") => "."
		// so we unify '.' => '' to match the listing page
		pDir := filepath.Dir(p.RelPath)
		if pDir == "." {
			pDir = ""
		}

		// if p is in the same dir, not an index, and has a date => sibling
		if pDir == dir && !p.IsIndex && !p.FrontMatter.ParsedDate.IsZero() {
			siblings = append(siblings, p)
		}
		// else fmt.Printf("[DEBUG] not-sibling: %s => pDir=%q\n", p.RelPath, pDir)
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
				link := "/" + outDir + "/"
				sb.WriteString(fmt.Sprintf(
					`<li><a class="text-decoration-none" href="%s">%s</a></li>`+"\n",
					link, art.FrontMatter.Title))
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
