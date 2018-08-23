package downloader

type Preparer interface {
	PrepareForCrawl() error
}

// Downloader a interface for download
type Downloader interface {
	Download(url string) ([]byte, error)
}
