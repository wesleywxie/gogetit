package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"github.com/wesleywxie/gogetit/internal/model"
)

func main() {
	url := "https://javdb.com/censored"
	items := []model.Item{}
	// Instantiate default collector
	collector := colly.NewCollector(
		// Visit only domains: reddit.com
		colly.AllowedDomains("javdb.com"),
		colly.MaxDepth(2), // only allow list and detail pages
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36"),
		colly.Debugger(&debug.LogDebugger{}),
	)
	detailCollector := collector.Clone()

	// On every a element which has .top-matter attribute call callback
	// This class is unique to the div that holds all information about a story
	collector.OnHTML(".grid-item", func(e *colly.HTMLElement) {
		temp := model.Item{}
		temp.UID = e.ChildText(".uid")
		temp.CrawledAt = time.Now()
		items = append(items, temp)
		url := e.Request.AbsoluteURL(e.ChildAttr("a[class=box]", "href"))
		detailCollector.Visit(url)
		detailCollector.Wait()
	})

	// On every span tag with the class next-button
	//collector.OnHTML("span.next-button", func(h *colly.HTMLElement) {
	//	t := h.ChildAttr("a", "href")
	//	collector.Visit(t)
	//})

	// Set max Parallelism and introduce a Random Delay
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*javdb.*",
		Parallelism: 1,
		Delay:       5 * time.Second,
	})

	// Before making a request print "Visiting ..."
	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Set error handler
	collector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	videos := []model.Video{}

	// On every a element which has .top-matter attribute call callback
	// This class is unique to the div that holds all information about a story
	detailCollector.OnHTML(".section .container", func(e *colly.HTMLElement) {
		temp := model.Video{}

		e.ForEach(".video-meta-panel .movie-panel-info .panel-block", func(_ int, el *colly.HTMLElement) {
			label := strings.TrimSpace(el.ChildText("strong:nth-child(1)"))
			switch label {
			case "番號:":
				temp.UID = el.ChildText("span")
			case "日期:":
				temp.PublishedAt = el.ChildText("span")
			case "時長:":
				temp.Duration = el.ChildText("span")
			case "導演:":
				temp.Director = el.ChildText("span")
			case "片商:":
				temp.Publisher = el.ChildText("span")
			case "系列:":
				temp.Series = el.ChildText("span")
			case "類別:":
				categories := strings.Split(el.ChildText("span"), ",")
				temp.Category = make([]string, len(categories))
				for i, category := range categories {
					temp.Category[i] = strings.TrimSpace(category)
				}
			case "演員:":
				//TODO how to extract male and female
				temp.Actress = strings.Fields(el.ChildText("span"))
			}
		})

		temp.Torrents = []model.Torrent{}
		reg := regexp.MustCompile(`\((.*?)\)`)

		e.ForEach("#magnets-content > table > tbody > tr", func(_ int, el *colly.HTMLElement) {
			t := model.Torrent{}
			t.Magnets = el.ChildAttr(".magnet-name > a", "href")
			metas := strings.Split(reg.FindAllString(el.ChildText(".meta"), -1)[0], ",")
			if len(metas) > 0 {
				t.Size = strings.TrimSpace(strings.Trim(strings.Trim(metas[0], "("), ")"))
				if len(metas) > 1 {
					t.Num = strings.TrimSpace(strings.Trim(strings.Trim(metas[1], "("), ")"))
				}
			}
			t.PublishedAt = el.ChildText(".time")
			temp.Torrents = append(temp.Torrents, t)
		})

		temp.Source = "JavDB"
		temp.CrawledAt = time.Now()
		videos = append(videos, temp)
	})

	// On every span tag with the class next-button
	//collector.OnHTML("span.next-button", func(h *colly.HTMLElement) {
	//	t := h.ChildAttr("a", "href")
	//	collector.Visit(t)
	//})

	// Set max Parallelism and introduce a Random Delay
	detailCollector.Limit(&colly.LimitRule{
		DomainGlob:  "*javdb.*",
		Parallelism: 1,
		Delay:       5 * time.Second,
	})

	// Before making a request print "Visiting ..."
	detailCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Set error handler
	detailCollector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	collector.Visit(url)
	collector.Wait()
	jsonData, _ := json.MarshalIndent(videos, "", "  ")
	fmt.Println(string(jsonData))

}
