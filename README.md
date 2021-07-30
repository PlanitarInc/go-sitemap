[![Go Reference](https://pkg.go.dev/badge/github.com/PlanitarInc/go-sitemap.svg)](https://pkg.go.dev/github.com/PlanitarInc/go-sitemap)
![CI Status](https://github.com/PlanitarInc/go-sitemap/actions/workflows/ci-flow.yml/badge.svg?branch=master)

A GO library for generating Sitemap XML files.

Example:

```go
package main

import (
	"os"
	"time"

	sitemap "github.com/PlanitarInc/go-sitemap"
)

type ArrayInput struct {
	Arr     []SimpleEntry
	NextIdx int
}

func (a ArrayInput) HasNext() bool {
	return a.NextIdx < len(a.Arr)
}

func (a *ArrayInput) Next() sitemap.UrlEntry {
	idx := a.NextIdx
	a.NextIdx++
	return a.Arr[idx]
}

type SimpleEntry struct {
	Url       string
	Modified  time.Time
	ImageUrls []string
}

func (e SimpleEntry) GetLoc() string {
	return e.Url
}

func (e SimpleEntry) GetLastMod() time.Time {
	return e.Modified
}

func (e SimpleEntry) GetImages() []string {
	return e.ImageUrls
}

func main() {
	entries := []SimpleEntry{
		SimpleEntry{
			Url:      "http://example.com/",
			Modified: time.Now(),
		},
		SimpleEntry{
			Url: "http://example.com/test/",
			ImageUrls: []string{
				"http://example.com/test/1.jpg",
				"http://example.com/test/2.jpg",
				"http://example.com/test/3.jpg",
			},
		},
	}

	sitemap.SitemapWrite(os.Stdout, &ArrayInput{Arr: entries})
}
```

The output:
```
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
  <url>
    <loc>http://example.com/</loc>
    <lastmod>2015-06-04T16:30:42-04:00</lastmod>
  </url>
  <url>
    <loc>http://example.com/test/</loc>
    <image:image>
      <image:loc>http://example.com/test/1.jpg</image:loc>
    </image:image>
    <image:image>
      <image:loc>http://example.com/test/2.jpg</image:loc>
    </image:image>
    <image:image>
      <image:loc>http://example.com/test/3.jpg</image:loc>
    </image:image>
  </url>
</urlset>
```
