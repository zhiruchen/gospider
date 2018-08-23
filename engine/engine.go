package engine

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"go.uber.org/zap"

	"github.com/zhiruchen/gospider/contentparser"
	"github.com/zhiruchen/gospider/downloader"
	"github.com/zhiruchen/gospider/log"
	"github.com/zhiruchen/gospider/proxy"
)

var (
	defaultReqTimeout       = 1 * time.Second
	defaultRetryTimes       = 3
	defaultReqInterval      = 100 * time.Millisecond
	defaultConcurrentReqNum = 100
)

const (
	urlQueueName = "seed_urls"
)

type Options struct {
	Headers          map[string]string
	Cookies          map[string]interface{}
	ReqTimeout       time.Duration
	RetryTimes       int
	ReqInterval      time.Duration // 单个请求之间的间隔
	ConcurrentReqNum int
}

var defaultServerOptions = Options{
	ReqTimeout:       defaultReqTimeout,
	RetryTimes:       defaultRetryTimes,
	ReqInterval:      defaultReqInterval,
	ConcurrentReqNum: defaultConcurrentReqNum,
}

type EngineOption func(o *Options)

func ReqHeaders(headers map[string]string) EngineOption {
	return func(o *Options) {
		o.Headers = headers
	}
}

func ReqCookies(cks map[string]interface{}) EngineOption {
	return func(o *Options) {
		o.Cookies = cks
	}
}

func ReqTimeout(t time.Duration) EngineOption {
	return func(o *Options) {
		o.ReqTimeout = t
	}
}

func RetryTimes(n int) EngineOption {
	return func(o *Options) {
		o.RetryTimes = n
	}
}

func ReqInterval(d time.Duration) EngineOption {
	return func(o *Options) {
		o.ReqInterval = d
	}
}

func ConcurrentReqNum(n int) EngineOption {
	return func(o *Options) {
		o.ConcurrentReqNum = n
	}
}

// Engine spider engine
type Engine struct {
	Name string

	opts *Options

	redisClient       *redis.Client
	startURLs         []string
	queueName         string
	urlSetName        string
	proxyURLQueueName string
	urlChs            chan string

	reqer  downloader.Requester
	prep   downloader.Preparer
	proxy  proxy.SpiderProxy
	dl     downloader.Downloader
	parser contentparser.Parser

	stats *Stats

	quit chan struct{}
	done chan struct{}
}

func NewEngine(redisClient *redis.Client, name string, urls []string, opts ...EngineOption) *Engine {
	defaultOpts := &defaultServerOptions
	for _, o := range opts {
		o(defaultOpts)
	}

	e := &Engine{
		Name:        name,
		opts:        defaultOpts,
		redisClient: redisClient,
		startURLs:   urls,
		urlChs:      make(chan string, defaultOpts.ConcurrentReqNum),
		queueName:   name + ":" + urlQueueName,
		urlSetName:  name + ":urls:set",
		stats:       newStats(),
		quit:        make(chan struct{}),
		done:        make(chan struct{}),
	}

	return e
}

func (e *Engine) SetProxy(pry proxy.SpiderProxy) {
	e.proxy = pry
}

func (e *Engine) SetRequester(reqer downloader.Requester) {
	e.reqer = reqer
}

func (e *Engine) SetDownloader(dl downloader.Downloader) {
	e.dl = dl
}

func (e *Engine) SetContentParser(parser contentparser.Parser) {
	e.parser = parser
}

func (e *Engine) SetPrepare(prep downloader.Preparer) {
	e.prep = prep
}

// Start start crawl
func (e *Engine) Start() {
	if err := e.prep.PrepareForCrawl(); err != nil {
		log.Logger.Error("prepare for crawl err", zap.Error(err))
		return
	}

	if err := e.enQueueURLs(e.startURLs); err != nil {
		log.Logger.Error("enqueue start urls err", zap.Error(err))
		return
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	go e.fillURLsChannelLoop()

	ticker := time.NewTicker(e.opts.ReqInterval)
Loop:
	for {
		select {
		case <-ticker.C:
			go e.processCrawl()
		case <-termChan:
			e.quit <- struct{}{}
			ticker.Stop()
			break Loop
		default:
		}
	}
}

func (e *Engine) processCrawl() {
Loop:
	for i := 1; i <= e.opts.ConcurrentReqNum; i++ {
		select {
		case v, ok := <-e.urlChs:
			if !ok {
				continue Loop
			}
			go e.crawl(v)
		default:
		}
	}
}

func (e *Engine) crawl(url string) {
	e.stats.IncrCrawlingCount(1)
	defer e.stats.IncrCrawlingCount(-1)

	var err error
	defer func() {
		if err != nil {
			e.stats.IncrFailedCount(1)
			return
		}
		e.stats.IncrCrawledCount(1)
	}()

	var content []byte
	content, err = e.Download(url)
	if err != nil {
		log.Logger.Error("download content err", zap.String("url", url), zap.Error(err))
		return
	}

	targetURLs, err := e.Parse(content)
	if err != nil {
		log.Logger.Error("parse content err", zap.String("url", url), zap.Error(err))
		return
	}

	if err = e.enQueueURLs(targetURLs); err != nil {
		log.Logger.Error("enqueue urls err", zap.Error(err))
	}

	if err = e.Save(); err != nil {
		log.Logger.Error("save result err", zap.String("url", url), zap.Error(err))
		return
	}
}

func (e *Engine) Stop() {
	close(e.urlChs)
	e.quit <- struct{}{}
}
