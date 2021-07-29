package sitemap

import "time"

type Input interface {
	Next() UrlEntry
	HasNext() bool
	SetIndexUrl(baseUrl string, fileName string, extension string)
	GetIndexUrl(idx int) string
}

type UrlEntry interface {
	GetLoc() string
	GetLastMod() time.Time
	GetImages() []string
}
