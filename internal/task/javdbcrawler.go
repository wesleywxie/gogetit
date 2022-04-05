package task

import (
	"fmt"
	"github.com/gocolly/colly"
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
	// Instantiate default collector
	collector := colly.NewCollector(
		// Visit only domains: reddit.com
		colly.AllowedDomains("javdb.com"),
		colly.MaxDepth(2), // only allow list and detail pages
		colly.Async(true),
		colly.UserAgent(config.UserAgent),
		colly.Debugger(&log.Debugger{}),
	)

	// Set max Parallelism and introduce a Random Delay
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*javdb.*",
		Parallelism: 1,
		Delay:       5 * time.Second,
	})

	if config.Socks5 != "" {
		err := collector.SetProxy(fmt.Sprintf("socks5://%s", config.Socks5))
		if err != nil {
			zap.S().Fatalw("Error when initializing proxy",
				"error", err,
			)
			return
		}
	}

	detailCollector := collector.Clone()

	collector.OnHTML(".grid-item", func(e *colly.HTMLElement) {
		UID := e.ChildText(".uid")
		URL := e.Request.AbsoluteURL(e.ChildAttr("a[class=box]", "href"))

		if !model.ExistsVideo(UID) {
			detailCollector.Visit(URL)
			detailCollector.Wait()
		}
	})

	// Before making a request print "Visiting ..."
	collector.OnRequest(func(r *colly.Request) {
		if t.isStop.Load() == true {
			r.Abort()
		}
	})

	// Set error handler
	collector.OnError(func(r *colly.Response, err error) {
		zap.S().Errorf("Request URL: %s failed with response: %v\nError:%v", r.Request.URL, r, err)
	})

	detailCollector.OnHTML(".section .container", func(e *colly.HTMLElement) {
		video := model.Video{}

		e.ForEach(".video-meta-panel .movie-panel-info .panel-block", func(_ int, el *colly.HTMLElement) {
			label := strings.TrimSpace(el.ChildText("strong:nth-child(1)"))
			switch label {
			case "番號:":
				video.UID = el.ChildText("span")
			case "日期:":
				video.PublishedAt = el.ChildText("span")
			case "時長:":
				video.Duration = el.ChildText("span")
			case "導演:":
				video.Director = el.ChildText("span")
			case "片商:":
				video.Publisher = el.ChildText("span")
			case "系列:":
				video.Series = el.ChildText("span")
			case "類別:":
				video.Categories = el.ChildText("span")
			case "演員:":
				video.Actors = el.ChildText("span")
			}
		})

		video.Source = "JavDB"
		video, _ = model.AddVideo(&video)

		reg := regexp.MustCompile(`\((.*?)\)`)

		e.ForEach("#magnets-content > table > tbody > tr", func(_ int, el *colly.HTMLElement) {
			t := model.Torrent{}
			t.Magnet = el.ChildAttr(".magnet-name > a", "href")
			metas := strings.Split(reg.FindAllString(el.ChildText(".meta"), -1)[0], ",")
			if len(metas) > 0 {
				t.Size = strings.TrimSpace(strings.Trim(strings.Trim(metas[0], "("), ")"))
				if len(metas) > 1 {
					t.Num = strings.TrimSpace(strings.Trim(strings.Trim(metas[1], "("), ")"))
				}
			}
			t.PublishedAt = el.ChildText(".time")
			t.VideoID = video.ID
			model.AddTorrent(&t)
		})

	})

	// Before making a request print "Visiting ..."
	detailCollector.OnRequest(func(r *colly.Request) {
		if t.isStop.Load() == true {
			r.Abort()
		}
	})

	detailCollector.OnError(func(r *colly.Response, err error) {
		zap.S().Errorf("Request URL: %s failed with response: %v\nError:%v", r.Request.URL, r, err)
	})

	collector.Visit(url)
	collector.Wait()
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
