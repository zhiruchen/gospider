package downloader

type Preparer interface {
	PrepareForCrawl() error
}

// Downloader a interface for download
type Downloader interface {
	Download(req Requester, url string) ([]byte, error)
}
