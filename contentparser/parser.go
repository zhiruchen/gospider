package contentparser

import (
	"github.com/zhiruchen/gospider/store"
)

type Parser interface {
	Parse(content []byte) ([]string, error)
	GetTargetURLs() []string
	store.Storage
}
