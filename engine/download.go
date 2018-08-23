package engine

import (
	"time"

	goreq "github.com/parnurzeal/gorequest"

	"github.com/zhiruchen/gospider/proxy"
)

var defaultRetryTime = 500 * time.Millisecond

type defaultDownloader struct {
	pry        proxy.SpiderProxy
	retryTimes int
	retryTime  time.Duration
}

func newDefaultDownloader(pry proxy.SpiderProxy, retryTimes int, retryTime time.Duration) *defaultDownloader {
	dl := &defaultDownloader{
		pry:        pry,
		retryTimes: retryTimes,
		retryTime:  retryTime,
	}

	if retryTimes > 1 && retryTime == 0 {
		dl.retryTime = defaultRetryTime
	}

	return dl
}

func (dl *defaultDownloader) Download(url string) ([]byte, error) {
	req := goreq.New().Proxy(dl.pry.GetProxyURL()).Retry(dl.retryTimes, dl.retryTime)
	_, body, errs := req.Get(url).End()

	if len(errs) > 0 {
		return nil, errs[0]
	}

	return []byte(body), nil
}

func (e *Engine) Download(url string) ([]byte, error) {
	return e.dl.Download(url)
}
