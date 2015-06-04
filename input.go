package sitemap

import "time"

type Input interface {
	Next() UrlEntry
	HasNext() bool
}

type UrlEntry interface {
	GetLoc() string
	GetLastMod() time.Time
	GetImages() []string
}
