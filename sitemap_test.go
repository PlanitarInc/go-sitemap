package sitemap

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

type ArrayInput struct {
	Arr     []SimpleEntry
	NextIdx int
}

func (a ArrayInput) HasNext() bool {
	return a.NextIdx < len(a.Arr)
}

func (a *ArrayInput) Next() UrlEntry {
	idx := a.NextIdx
	a.NextIdx++
	return a.Arr[idx]
}

type SimpleEntry struct {
	Loc     string
	LastMod time.Time
	Images  []string
}

func (e SimpleEntry) GetLoc() string {
	return e.Loc
}

func (e SimpleEntry) GetLastMod() time.Time {
	return e.LastMod
}

func (e SimpleEntry) GetImages() []string {
	return e.Images
}

func TestSitemapWriteEmpty(t *testing.T) {
	RegisterTestingT(t)

	var out bytes.Buffer

	Ω(SitemapWrite(&out, &ArrayInput{})).Should(BeNil())
	Ω(out.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1"></urlset>
	`)))
}

func TestSitemapWriteSimple(t *testing.T) {
	RegisterTestingT(t)

	var out bytes.Buffer
	var entries []SimpleEntry

	out.Reset()
	entries = []SimpleEntry{
		SimpleEntry{},
		SimpleEntry{},
	}
	Ω(SitemapWrite(&out, &ArrayInput{Arr: entries})).Should(BeNil())
	Ω(out.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
  <url>
    <loc></loc>
  </url>
  <url>
    <loc></loc>
  </url>
</urlset>
	`)))

	out.Reset()
	entries = []SimpleEntry{
		SimpleEntry{
			Loc:     "one",
			LastMod: time.Date(1999, 12, 31, 23, 59, 59, 0, time.UTC),
		},
		SimpleEntry{
			Loc:     "two",
			LastMod: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		SimpleEntry{
			Loc:     "three",
			LastMod: time.Date(2015, 7, 22, 15, 48, 2, 0, time.UTC),
		},
	}
	Ω(SitemapWrite(&out, &ArrayInput{Arr: entries})).Should(BeNil())
	Ω(out.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
  <url>
    <loc>one</loc>
  </url>
  <url>
    <loc>two</loc>
    <lastmod>2001-01-01T00:00:00Z</lastmod>
  </url>
  <url>
    <loc>three</loc>
    <lastmod>2015-07-22T15:48:02Z</lastmod>
  </url>
</urlset>
	`)))
}

func TestSitemapWriteImages(t *testing.T) {
	RegisterTestingT(t)

	var out bytes.Buffer
	var entries []SimpleEntry

	out.Reset()
	entries = []SimpleEntry{
		SimpleEntry{
			Images: []string{},
		},
		SimpleEntry{},
	}
	Ω(SitemapWrite(&out, &ArrayInput{Arr: entries})).Should(BeNil())
	Ω(out.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
  <url>
    <loc></loc>
  </url>
  <url>
    <loc></loc>
  </url>
</urlset>
	`)))

	out.Reset()
	entries = []SimpleEntry{
		SimpleEntry{
			Loc:    "one",
			Images: []string{"a", "b", "c"},
		},
		SimpleEntry{
			Loc: "two",
		},
		SimpleEntry{
			Loc:    "three",
			Images: []string{"w", "x", "y", "z"},
		},
	}
	Ω(SitemapWrite(&out, &ArrayInput{Arr: entries})).Should(BeNil())
	Ω(out.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
  <url>
    <loc>one</loc>
    <image:image>
      <image:loc>a</image:loc>
    </image:image>
    <image:image>
      <image:loc>b</image:loc>
    </image:image>
    <image:image>
      <image:loc>c</image:loc>
    </image:image>
  </url>
  <url>
    <loc>two</loc>
  </url>
  <url>
    <loc>three</loc>
    <image:image>
      <image:loc>w</image:loc>
    </image:image>
    <image:image>
      <image:loc>x</image:loc>
    </image:image>
    <image:image>
      <image:loc>y</image:loc>
    </image:image>
    <image:image>
      <image:loc>z</image:loc>
    </image:image>
  </url>
</urlset>
	`)))
}

type DynamicInput struct {
	Size    int
	NextIdx int
	Entry   SimpleEntry
}

func (d DynamicInput) HasNext() bool {
	return d.NextIdx < d.Size
}

func (d *DynamicInput) Next() UrlEntry {
	d.NextIdx++
	return d.Entry
}

func benchSitemap(size int, b *testing.B) {
	in := DynamicInput{
		Size: size,
		Entry: SimpleEntry{
			Loc:     "http://www.example.com/qweqwe",
			LastMod: minDate.AddDate(1, 2, 3),
			Images:  []string{"http://www.example.com/qweqwe/thumb.jpg"},
		},
	}

	for n := 0; n < b.N; n++ {
		SitemapWrite(ioutil.Discard, &in)
	}
}

func BenchmarkSitemap10(b *testing.B)   { benchSitemap(10, b) }
func BenchmarkSitemap100(b *testing.B)  { benchSitemap(100, b) }
func BenchmarkSitemap1K(b *testing.B)   { benchSitemap(1000, b) }
func BenchmarkSitemap10K(b *testing.B)  { benchSitemap(10000, b) }
func BenchmarkSitemap100K(b *testing.B) { benchSitemap(100000, b) }
func BenchmarkSitemap1M(b *testing.B)   { benchSitemap(1000000, b) }
