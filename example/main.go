package main

import (
	"time"
	"bytes"
	"fmt"

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
	var output []bytes.Buffer
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

	sitemap.SitemapWrite(&output, &ArrayInput{Arr: entries})
	fmt.Println(output[0].String())
}
