package sitemap_test

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/PlanitarInc/go-sitemap"
)

type SiteMapOutlet struct {
	indexBuf   bytes.Buffer
	siteMapBuf []bytes.Buffer
}

func (out *SiteMapOutlet) Index() io.Writer {
	return &out.indexBuf
}

func (out *SiteMapOutlet) Urlset() io.Writer {
	out.siteMapBuf = append(out.siteMapBuf, bytes.Buffer{})
	return &out.siteMapBuf[len(out.siteMapBuf)-1]
}

func ExampleWriteWithIndex() {
	var out SiteMapOutlet
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
	// Output: <?xml version="1.0" encoding="UTF-8"?>
	// <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
	//   <url>
	//     <loc>http://example.com/</loc>
	//     <lastmod>2025-11-02T11:34:58Z</lastmod>
	//   </url>
	//   <url>
	//     <loc>http://example.com/test/</loc>
	//     <image:image>
	//       <image:loc>http://example.com/test/1.jpg</image:loc>
	//     </image:image>
	//     <image:image>
	//       <image:loc>http://example.com/test/2.jpg</image:loc>
	//     </image:image>
	//     <image:image>
	//       <image:loc>http://example.com/test/3.jpg</image:loc>
	//     </image:image>
	//   </url>
	// </urlset>
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
