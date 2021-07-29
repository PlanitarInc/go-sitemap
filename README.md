[![Go Reference](https://pkg.go.dev/badge/github.com/PlanitarInc/go-sitemap.svg)](https://pkg.go.dev/github.com/PlanitarInc/go-sitemap)
![CI Status](https://github.com/PlanitarInc/go-sitemap/actions/workflows/ci-flow.yml/badge.svg?branch=master)

A GO library for generation of sitemap XML files.

Example:

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/PlanitarInc/go-sitemap"
)

type SiteMapOutput struct {
	indexBuf   bytes.Buffer
	siteMapBuf []bytes.Buffer
}

func (out *SiteMapOutput) Index() io.Writer {
	return &out.indexBuf
}

func (out *SiteMapOutput) Urlset() io.Writer {
	out.siteMapBuf = append(out.siteMapBuf, bytes.Buffer{})
	return &out.siteMapBuf[len(out.siteMapBuf)-1]
}

type ArrayInput struct {
	Arr       []SimpleEntry
	NextIdx   int
	baseUrl   string
	fileName   string
	extension string
}

func (a ArrayInput) HasNext() bool {
	return a.NextIdx < len(a.Arr)
}

func (a *ArrayInput) Next() sitemap.UrlEntry {
	idx := a.NextIdx
	a.NextIdx++
	return a.Arr[idx]
}

func (a *ArrayInput) SetIndexUrl(baseUrl string, fileName string, extension string) {
	a.baseUrl = baseUrl
	a.fileName = fileName
	a.extension = extension
}

func (a *ArrayInput) GetIndexUrl(idx int) string {
	return a.baseUrl + a.fileName + strconv.Itoa(idx+1) + "." + a.extension
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

func (a *ArrayInput) GetUrlsetUrl(idx int) string {
	return ""
}

func main() {
	var out SiteMapOutput
	entries := []SimpleEntry{
		{
			Url:      "http://example.com/",
			Modified: time.Date(2025, time.November, 2, 11, 34, 58, 123, time.UTC),
		},
		{
			Url: "http://example.com/test/",
			ImageUrls: []string{
				"http://example.com/test/1.jpg",
				"http://example.com/test/2.jpg",
				"http://example.com/test/3.jpg",
			},
		},
	}

	_ = sitemap.WriteWithIndex(&out, &ArrayInput{Arr: entries}, 5)
	fmt.Println(out.siteMapBuf[0].String())
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
