package engine

import (
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/zhiruchen/gospider/log"
)

var cutsets = []string{
	" ",
	"\n",
	"\t",
}

func (e *Engine) enQueueURLs(urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	for i := range urls {
		for _, cutset := range cutsets {
			urls[i] = strings.Trim(urls[i], cutset)
		}
	}

	var newURLs []interface{}
	for _, url := range urls {
		v, err := e.redisClient.SAdd(e.urlSetName, url).Result()
		if err != nil {
			return err
		}

		// url 不存在则说明未爬过
		if v == 1 {
			newURLs = append(newURLs, url)
		}
	}

	if _, err := e.redisClient.RPush(e.queueName, newURLs...).Result(); err != nil {
		return err
	}

	return nil
}

// fillURLChan 定时从redis拿url放到engine的url通道中
func (e *Engine) fillURLsChannelLoop() {
	ticket := time.NewTicker(500 * time.Millisecond)
Loop:
	for {
		select {
		case <-ticket.C:
		Inner:
			for i := 1; i <= e.opts.ConcurrentReqNum; i++ {
				url, err := e.redisClient.LPop(e.queueName).Result()
				if err != nil {
					log.Logger.Error("redis lpop url err", zap.Error(err))
					continue Loop
				}

				if url == "" {
					continue Inner
				}

				select {
				case e.urlChs <- url:
				default:
					continue Loop // urlChs 已满
				}
			}
		case <-e.quit:
			ticket.Stop()
			break Loop
		}
	}
}
