package sitemap

import (
	"io"
	"time"
)

type Input interface {
	Next() UrlEntry
	HasNext() bool
	GetUrlsetUrl(idx int) string
}

type UrlEntry interface {
	GetLoc() string
	GetLastMod() time.Time
	GetImages() []string
}

type Output interface {
	Index() io.Writer
	Urlset() io.Writer
}
