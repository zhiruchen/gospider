package engine

import (
	"sync/atomic"
)

// Stats engine stats
type Stats struct {
	CrawlingCount int64
	CrawledCount  int64
	FailedCount   int64
}

func newStats() *Stats {
	return &Stats{
		CrawlingCount: 0,
		CrawledCount:  0,
		FailedCount:   0,
	}
}

func (s *Stats) IncrCrawlingCount(v int64) {
	atomic.AddInt64(&s.CrawlingCount, v)
}

func (s *Stats) IncrCrawledCount(v int64) {
	atomic.AddInt64(&s.CrawledCount, v)
}

func (s *Stats) IncrFailedCount(v int64) {
	atomic.AddInt64(&s.FailedCount, v)
}
