package sitemap

import (
	"io"
	"time"
)

type Input interface {
	Next() *UrlEntry
	HasNext() bool
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
