package sitemap

import (
	"io"
	"time"
)

type Input interface {
	// Next returns the next UrlEntry to be written. The function should
	// return nil if and only if there are no more items.
	Next() *UrlEntry
	// GetUrlsetUrl returns a URL for the Urlset file at the given index.
	GetUrlsetUrl(idx int) string
}

type UrlEntry struct {
	Loc     string
	LastMod time.Time
	Images  []string
}

type Output interface {
	Index() io.Writer
	Urlset() io.Writer
}
