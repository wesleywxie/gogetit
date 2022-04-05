package task

import (
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/proxy"
	"github.com/wesleywxie/gogetit/internal/config"
	"github.com/wesleywxie/gogetit/internal/log"
	"github.com/wesleywxie/gogetit/internal/model"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func init() {
	task := NewJavDBCrawler()
	registerTask(task)
}

// JavDBCrawler 爬取最新JavDB种子
type JavDBCrawler struct {
	isStop atomic.Bool
}

// NewJavDBCrawler 构造 JavDBCrawler
func NewJavDBCrawler() *JavDBCrawler {
	task := &JavDBCrawler{}
	task.isStop.Store(false)
	return task
}

// Name 任务名称
func (t *JavDBCrawler) Name() string {
	return "JavDBCrawler"
}

// Stop 停止
func (t *JavDBCrawler) Stop() {
	t.isStop.Store(true)
}

// Start 启动
func (t *JavDBCrawler) Start() {
	t.isStop.Store(false)

	url := "https://javdb.com/censored"
	items := []model.Item{}
	count := 0
	// Instantiate default collector
	collector := colly.NewCollector(
		// Visit only domains: reddit.com
		colly.AllowedDomains("javdb.com"),
		colly.MaxDepth(2), // only allow list and detail pages
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36"),
		colly.Debugger(&log.Debugger{}),
	)

	if config.Socks5 != "" {
		rp, err := proxy.RoundRobinProxySwitcher(fmt.Sprintf("socks5://%s", config.Socks5))
		if err != nil {
			zap.S().Fatalw("Error when initializing proxy",
				"error", err,
			)
		}

		collector.SetProxyFunc(rp)
	}

	detailCollector := collector.Clone()

	// On every a element which has .top-matter attribute call callback
	// This class is unique to the div that holds all information about a story
	collector.OnHTML(".grid-item", func(e *colly.HTMLElement) {
		temp := model.Item{}
		temp.UID = e.ChildText(".uid")
		temp.URL = e.Request.AbsoluteURL(e.ChildAttr("a[class=box]", "href"))
		temp.CrawledAt = time.Now()
		items = append(items, temp)

		for ; count < 5; count++ {
			detailCollector.Visit(temp.URL)
			detailCollector.Wait()
		}
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
		zap.S().Debugf("Visiting %s", r.URL.String())
	})

	// Set error handler
	collector.OnError(func(r *colly.Response, err error) {
		zap.S().Errorf("Request URL: %s failed with response: %v\nError:%v", r.Request.URL, r, err)
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
				temp.Categories = el.ChildText("span")
			case "演員:":
				temp.Actors = el.ChildText("span")
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
			t.CreatedAt = time.Now()
			t.UpdatedAt = time.Now()
			temp.Torrents = append(temp.Torrents, t)
		})

		temp.Source = "JavDB"
		temp.CreatedAt = time.Now()
		temp.UpdatedAt = time.Now()
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
		zap.S().Debugf("Visiting %s", r.URL.String())
	})

	// Set error handler
	detailCollector.OnError(func(r *colly.Response, err error) {
		zap.S().Errorf("Request URL: %s failed with response: %v\nError:%v", r.Request.URL, r, err)
	})

	collector.Visit(url)
	collector.Wait()
	jsonData, _ := json.MarshalIndent(videos, "", "  ")
	zap.S().Debug(string(jsonData))
}

func makeGetRequest(url string) (content string, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		zap.S().Error(err)
		return
	}
	req.Header.Set("User-Agent", config.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		zap.S().Error(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		zap.S().Error(err)
		return
	}

	content = string(body)
	return
}
