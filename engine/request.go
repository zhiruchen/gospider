package engine

import (
	"time"
)

// Status request status
type Status string

const (
	Timeout           Status = "Timeout"
	NotFound          Status = "NotFound"
	InternalServerErr Status = "InternalServerErr"
)

// Request crawl web request status
type TaskStatus struct {
	URL        string    `json:"url"`
	CreatedAt  time.Time `json:"created_at"`
	FinishedAt time.Time `json:"finished_at"`
	Status     Status    `json:"status"`
}

func (s *TaskStatus) Sync() {

}
