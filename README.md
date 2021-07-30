[![Go Reference](https://pkg.go.dev/badge/github.com/PlanitarInc/go-sitemap.svg)](https://pkg.go.dev/github.com/PlanitarInc/go-sitemap)
![CI Status](https://github.com/PlanitarInc/go-sitemap/actions/workflows/ci-flow.yml/badge.svg?branch=master)

A GO library for generating Sitemap XML files.

Example:

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"time"

	sitemap "github.com/PlanitarInc/go-sitemap"
)

type SitemapOutput struct {
	IndexBuf   bytes.Buffer
	UrlsetBufs []bytes.Buffer
}

func (out *SitemapOutput) Index() io.Writer {
	return &out.IndexBuf
}

func (out *SitemapOutput) Urlset() io.Writer {
	out.UrlsetBufs = append(out.UrlsetBufs, bytes.Buffer{})
	return &out.UrlsetBufs[len(out.UrlsetBufs)-1]
}

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

func (a *ArrayInput) GetUrlsetUrl(n int) string {
	return fmt.Sprintf("https://example.com/sitemap-%d.xml", n)
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

	var out SitemapOutput
	err := sitemap.WriteAll(&out, &ArrayInput{Arr: entries})
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	for i := range out.UrlsetBufs {
		fmt.Printf("\n\n::: Urlset %d\n\n", i)
		fmt.Print(out.UrlsetBufs[i].String())
	}
	fmt.Printf("\n\n::: Index\n\n")
	fmt.Print(out.IndexBuf.String())
}
```

The output:
```
::: Urlset 0

<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
  <url>
    <loc>http://example.com/</loc>
    <lastmod>2025-11-02T11:34:58Z</lastmod>
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

::: Index

<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/sitemap-0.xml</loc>
  </url>
</sitemapindex>
```
