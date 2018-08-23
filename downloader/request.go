package downloader

import (
	"github.com/zhiruchen/gospider/proxy"
)

type Requester interface {
	proxy.SpiderProxy
	Get(url string) ([]byte, error)
}
