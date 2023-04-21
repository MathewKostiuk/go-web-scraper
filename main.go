package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/tealeg/xlsx"
)

type Section struct {
	Name   string
	Count  int
	Unused bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <website>")
		os.Exit(1)
	}

	websiteArg := os.Args[1]
	parsedURL, err := url.Parse(websiteArg)
	if err != nil {
		log.Fatal("Error parsing website URL:", err)
	}

	website := parsedURL.Hostname()
	targetURL := parsedURL.Scheme + "://" + website

	sectionDir := "./sections"

	files, err := ioutil.ReadDir(sectionDir)
	if err != nil {
		log.Fatal("Error reading sections directory:", err)
	}

	sectionNames := make([]string, len(files))
	for i, file := range files {
		sectionNames[i] = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
	}

	c := colly.NewCollector(
		colly.AllowedDomains(website, "www."+website),
		colly.MaxDepth(3),
	)

	sections := make(map[string]*Section)
	for _, name := range sectionNames {
		sections[name] = &Section{
			Name:   name,
			Count:  0,
			Unused: true,
		}
	}

	c.OnHTML("*", func(e *colly.HTMLElement) {
		sectionFound := ""
		for _, name := range sectionNames {
			if strings.Contains(e.Attr("class"), name) || e.Attr("data-section") == name || strings.HasPrefix(e.Attr("id"), "shopify-section-"+name) {
				sectionFound = name
				break
			}
		}

		if sectionFound != "" {
			sections[sectionFound].Count++
			sections[sectionFound].Unused = false
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

		if parsedLink.Hostname() == website {
			visitedMutex.Lock()

			if _, exists := visited[link]; !exists {
				visited[link] = struct{}{}
				visitedMutex.Unlock()

				err := c.Visit(link)
				if err != nil {
					log.Printf("Error visiting link %s: %v", link, err)
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
	row.AddCell().Value = "Unused"
	for _, section := range sortedSections {
		row := sheet.AddRow()
		row.AddCell().Value = section.Name
		row.AddCell().SetInt(section.Count)
		row.AddCell().SetBool(section.Unused)
	}

	err = file.Save("shopify_sections.xlsx")
	if err != nil {
		log.Fatal("Error saving xlsx file:", err)
	}

	fmt.Println("Scraping and ranking complete. Saved to shopify_sections.xlsx.")
}
