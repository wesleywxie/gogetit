package task

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/wesleywxie/gogetit/internal/config"
	"github.com/wesleywxie/gogetit/internal/log"
	"github.com/wesleywxie/gogetit/internal/model"
	"github.com/wesleywxie/gogetit/internal/util"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
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
func (t *JavDBCrawler) IsStopped() bool {
	return t.isStop.Load()
}

// Stop 停止
func (t *JavDBCrawler) Stop() {
	t.isStop.Store(true)
}

// Start 启动
func (t *JavDBCrawler) Start() {
	t.isStop.Store(false)

	go func() {
		for {
			if t.isStop.Load() == true {
				zap.S().Infof("%s stopped", t.Name())
				return
			}
			// Instantiate default collector
			collector := createCollector()
			detailCollector := collector.Clone()

			// Before making a request print "Visiting ..."
			collector.OnRequest(func(r *colly.Request) {
				if t.isStop.Load() == true {
					r.Abort()
				}
			})

			collector.OnHTML(".grid-item", func(e *colly.HTMLElement) {
				UID := e.ChildText(".uid")
				URL := e.Request.AbsoluteURL(e.ChildAttr("a[class=box]", "href"))

				if !model.ExistsVideo(UID) {
					_ = detailCollector.Visit(URL)
					detailCollector.Wait()
				}
			})

			// Set error handler
			collector.OnError(func(r *colly.Response, err error) {
				zap.S().Errorf("Request URL: %s failed with response: %v\nError:%v", r.Request.URL, r, err)
			})

			// Before making a request print "Visiting ..."
			detailCollector.OnRequest(func(r *colly.Request) {
				if t.isStop.Load() == true {
					r.Abort()
				}
			})

			detailCollector.OnHTML(".section .container", func(e *colly.HTMLElement) {
				video := parseVideo(e)
				parseTorrent(video, e)

			})
			detailCollector.OnError(func(r *colly.Response, err error) {
				zap.S().Errorf("Request URL: %s failed with response: %v\nError:%v", r.Request.URL, r, err)
			})

			url := "https://javdb.com/censored?page=%d"
			for i := 1; i < 6; i++ {
				_ = collector.Visit(fmt.Sprintf(url, i))
				collector.Wait()
			}

			time.Sleep(time.Duration(config.UpdateInterval) * time.Minute)
		}
	}()
}

func createCollector() *colly.Collector {
	// Instantiate default collector
	collector := colly.NewCollector(
		// Visit only domains: reddit.com
		colly.AllowedDomains("javdb.com"),
		colly.MaxDepth(2), // only allow list and detail pages
		colly.Async(false),
		colly.UserAgent(config.UserAgent),
		colly.Debugger(&log.Debugger{}),
	)

	// Set max Parallelism and introduce a Random Delay
	_ = collector.Limit(&colly.LimitRule{
		DomainGlob:  "*javdb.*",
		Parallelism: 1,
		Delay:       2 * time.Second,
	})

	if len(config.Socks5) > 0 {
		dialContext, err := newDialContext(config.Socks5)
		if err != nil {
			zap.S().Errorw("Error when initializing proxy", "error", err)
			return nil
		}

		collector.WithTransport(&http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		})
	}

	return collector
}

func parseVideo(e *colly.HTMLElement) model.Video {
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
	return video
}

func parseTorrent(video model.Video, e *colly.HTMLElement) {
	e.ForEach("#magnets-content > .item", func(_ int, el *colly.HTMLElement) {
		t := model.Torrent{}
		t.MagnetLink = strings.Split(el.ChildAttr(".magnet-name > a", "href"), "&")[0]
		metas := strings.Split(el.ChildText(".meta"), ",")
		if len(metas) > 0 && len(metas[0]) > 0 {
			size := strings.TrimSpace(metas[0])
			unit := size[len(size)-2:]
			multiplier := 1
			if strings.ToUpper(unit) == "GB" {
				multiplier = 1024
			}
			t.FileSize = int(util.ExtractFloat(size) * float64(multiplier))

			if len(metas) > 1 {
				num := strings.TrimSpace(metas[1])
				t.FileNum = util.ExtractInt(num)
			}
		}
		t.PublishedAt, _ = util.ParseTime(el.ChildText(".time"))
		t.VideoID = video.ID
		t.UID = video.UID
		_, _ = model.AddTorrent(&t)
	})
}

type dialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

func newDialContext(socks5 string) (dialContextFunc, error) {
	baseDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	if socks5 != "" {
		dialSocksProxy, err := proxy.SOCKS5("tcp", socks5, nil, baseDialer)
		if err != nil {
			return nil, err
		}

		contextDialer, ok := dialSocksProxy.(proxy.ContextDialer)
		if !ok {
			return nil, err
		}

		return contextDialer.DialContext, nil
	} else {
		return baseDialer.DialContext, nil
	}
}
