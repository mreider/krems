package main

import (
	"fmt"
	"html/template"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// listPagesInDirectory => for type:list pages
func listPagesInDirectory(relPath string) template.HTML {
	if globalBuildCache == nil {
		return ""
	}
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

	dir := filepath.Dir(listingPage.RelPath)
	if dir == "." {
		dir = ""
	}

	// find siblings with valid date
	var siblings []*PageData
	for _, p := range globalBuildCache.Pages {
		if filepath.Dir(p.RelPath) == dir && !p.IsIndex && !p.FrontMatter.ParsedDate.IsZero() {
			siblings = append(siblings, p)
		}
	}
	sort.Slice(siblings, func(i, j int) bool {
		return siblings[i].FrontMatter.ParsedDate.After(siblings[j].FrontMatter.ParsedDate)
	})

	// group by year => month
	groups := groupByYearThenMonth(siblings)

	var sb strings.Builder
	sb.WriteString(`<div class="blog-list">`)
	for _, yg := range groups {
		sb.WriteString(fmt.Sprintf(`<h3 class="mt-5 mb-3">%d</h3>`+"\n", yg.Year))
		for _, mg := range yg.Months {
			sb.WriteString(fmt.Sprintf(`<h5 class="mb-2">%s</h5>`+"\n", mg.Month))
			sb.WriteString(`<ul class="list-group mb-4">` + "\n")
			for _, art := range mg.Pages {
				outDir := FindPageByRelPath(globalBuildCache, art.RelPath)
				link := "/" + outDir + "/"
				sb.WriteString(fmt.Sprintf(
					`<li class="list-group-item"><a class="text-decoration-none" href="%s">%s</a></li>`+"\n",
					link, art.FrontMatter.Title))
			}
			sb.WriteString("</ul>\n")
		}
	}
	sb.WriteString(`</div>`)
	return template.HTML(sb.String())
}

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
	var years []int
	for y := range yMap {
		years = append(years, y)
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

type yearGroup struct {
	Year   int
	Months []monthGroup
}
type monthGroup struct {
	Month time.Month
	Pages []*PageData
}
