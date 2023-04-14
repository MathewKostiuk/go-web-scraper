package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/tealeg/xlsx"
)

type Section struct {
	Name       string
	Count      int
	IsHomepage bool
	Unused     bool
}

func main() {
	sectionDir := "./sections"
	targetURL := "https://fentybeauty.com"

	files, err := ioutil.ReadDir(sectionDir)
	if err != nil {
		log.Fatal("Error reading sections directory:", err)
	}

	sectionNames := make([]string, len(files))
	for i, file := range files {
		sectionNames[i] = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
	}

	c := colly.NewCollector(
		colly.AllowedDomains("fentybeauty.com"),
		colly.MaxDepth(3),
	)

	sections := make(map[string]*Section)
	for _, name := range sectionNames {
		sections[name] = &Section{
			Name:       name,
			Count:      0,
			IsHomepage: false,
			Unused:     true,
		}
	}

	c.OnHTML("*", func(e *colly.HTMLElement) {
		sectionFound := ""
		for _, name := range sectionNames {
			if strings.Contains(e.Attr("class"), name) || e.Attr("data-module") == name || strings.HasPrefix(e.Attr("id"), "shopify-section-"+name) {
				sectionFound = name
				break
			}
		}

		if sectionFound != "" {
			sections[sectionFound].Count++
			sections[sectionFound].Unused = false

			if !sections[sectionFound].IsHomepage {
				sections[sectionFound].IsHomepage = strings.Contains(e.Request.URL.Path, "/")
			}
		}
	})

	visited := make(map[string]struct{})
	var visitedMutex sync.Mutex

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		parsedLink, err := url.Parse(link)
		if err != nil {
			return
		}

		if parsedLink.Hostname() == "fentybeauty.com" {
			visitedMutex.Lock()

			if _, exists := visited[link]; !exists {
				visited[link] = struct{}{}
				visitedMutex.Unlock()

				err := c.Visit(link)
				if err != nil {
					log.Println("Error visiting link:", err)
				}
			} else {
				visitedMutex.Unlock()
			}
		}
	})

	err = c.Visit(targetURL)
	if err != nil {
		log.Fatal("Error visiting target URL:", err)
	}

	sortedSections := make([]*Section, 0, len(sections))
	for _, section := range sections {
		sortedSections = append(sortedSections, section)
	}

	sort.Slice(sortedSections, func(i, j int) bool {
		return sortedSections[i].Count > sortedSections[j].Count
	})

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sections")
	if err != nil {
		log.Fatal("Error adding sheet:", err)
	}

	row := sheet.AddRow()

	row.AddCell().Value = "Section Name"
	row.AddCell().Value = "Usage Count"
	row.AddCell().Value = "Used on Homepage"
	row.AddCell().Value = "Unused"
	for _, section := range sortedSections {
		row := sheet.AddRow()
		row.AddCell().Value = section.Name
		row.AddCell().SetInt(section.Count)
		row.AddCell().SetBool(section.IsHomepage)
		row.AddCell().SetBool(section.Unused)
	}

	err = file.Save("shopify_sections_v4.xlsx")
	if err != nil {
		log.Fatal("Error saving xlsx file:", err)
	}

	fmt.Println("Scraping and ranking complete. Saved to shopify_sections_v4.xlsx.")
}
