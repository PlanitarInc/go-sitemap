package sitemap

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
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

type ArrayInput struct {
	Arr     []SimpleEntry
	NextIdx int
}

func (a ArrayInput) HasNext() bool {
	return a.NextIdx < len(a.Arr)
}

func (a *ArrayInput) GetUrlsetUrl(idx int) string {
	return "https://youriguide.com/sitemap/view" + strconv.Itoa(idx+1) + ".xml"
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

func TestWriteWithIndexEmpty(t *testing.T) {
	RegisterTestingT(t)

	var out SiteMapOutlet

	Ω(WriteWithIndex(&out, &ArrayInput{}, 5)).Should(BeNil())
	Ω(out.siteMapBuf[0].String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1"></urlset>
	`)))
}

func TestWriteWithIndexSimple(t *testing.T) {
	RegisterTestingT(t)

	var out SiteMapOutlet
	entries := []SimpleEntry{
		SimpleEntry{},
		SimpleEntry{},
	}
	Ω(WriteWithIndex(&out, &ArrayInput{Arr: entries}, 5)).Should(BeNil())
	Ω(out.siteMapBuf[0].String()).Should(Equal(strings.TrimSpace(`
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
}
func TestWriteWithIndexSimple2(t *testing.T) {
	RegisterTestingT(t)
	var out SiteMapOutlet
	entries := []SimpleEntry{
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
		SimpleEntry{
			Loc:     "four",
			LastMod: time.Date(1999, 12, 31, 23, 59, 59, 0, time.UTC),
		},
		SimpleEntry{
			Loc:     "five",
			LastMod: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		SimpleEntry{
			Loc:     "six",
			LastMod: time.Date(2015, 7, 22, 15, 48, 2, 0, time.UTC),
		},
	}
	Ω(WriteWithIndex(&out, &ArrayInput{Arr: entries}, 5)).Should(BeNil())
	Ω(out.siteMapBuf[0].String()).Should(Equal(strings.TrimSpace(`
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
  <url>
    <loc>four</loc>
  </url>
  <url>
    <loc>five</loc>
    <lastmod>2001-01-01T00:00:00Z</lastmod>
  </url>
</urlset>
	`)))
	Ω(out.siteMapBuf[1].String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
  <url>
    <loc>six</loc>
    <lastmod>2015-07-22T15:48:02Z</lastmod>
  </url>
</urlset>
	`)))
	Ω(out.indexBuf.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://youriguide.com/sitemap/view1.xml</loc>
  </url>
  <url>
    <loc>https://youriguide.com/sitemap/view2.xml</loc>
  </url>
</sitemapindex>
	`)))
}

func TestWriteWithIndexImages(t *testing.T) {
	RegisterTestingT(t)

	var out SiteMapOutlet
	entries := []SimpleEntry{
		SimpleEntry{
			Images: []string{},
		},
		SimpleEntry{},
	}
	Ω(WriteWithIndex(&out, &ArrayInput{Arr: entries}, 5)).Should(BeNil())
	Ω(out.siteMapBuf[0].String()).Should(Equal(strings.TrimSpace(`
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
}
func TestWriteWithIndexImages2(t *testing.T) {
	RegisterTestingT(t)

	var out SiteMapOutlet
	entries := []SimpleEntry{
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
	Ω(WriteWithIndex(&out, &ArrayInput{Arr: entries}, 5)).Should(BeNil())
	Ω(out.siteMapBuf[0].String()).Should(Equal(strings.TrimSpace(`
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

func TestWriteWithIndexEscaping(t *testing.T) {
	RegisterTestingT(t)

	var out SiteMapOutlet
	entries := []SimpleEntry{
		SimpleEntry{
			Loc:    `http://www.example.com/q="<'a'&'b'>"`,
			Images: []string{`"<`, `qwe&qw&ewq`, `asd`},
		},
	}
	Ω(WriteWithIndex(&out, &ArrayInput{Arr: entries}, 5)).Should(BeNil())
	Ω(out.siteMapBuf[0].String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
  <url>
    <loc>http://www.example.com/q=&#34;&lt;&#39;a&#39;&amp;&#39;b&#39;&gt;&#34;</loc>
    <image:image>
      <image:loc>&#34;&lt;</image:loc>
    </image:image>
    <image:image>
      <image:loc>qwe&amp;qw&amp;ewq</image:loc>
    </image:image>
    <image:image>
      <image:loc>asd</image:loc>
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

func (a *DynamicInput) GetUrlsetUrl(idx int) string {
	return ""
}

func benchSitemap(size int, b *testing.B) {
	var out SiteMapOutlet
	in := DynamicInput{
		Size: size,
		Entry: SimpleEntry{
			Loc:     "http://www.example.com/qweqwe",
			LastMod: minDate.AddDate(1, 2, 3),
			Images:  []string{"http://www.example.com/qweqwe/thumb.jpg"},
		},
	}

	for n := 0; n < b.N; n++ {
		_ = WriteWithIndex(&out, &in, 5)
	}
}

func BenchmarkSitemap10(b *testing.B)   { benchSitemap(10, b) }
func BenchmarkSitemap100(b *testing.B)  { benchSitemap(100, b) }
func BenchmarkSitemap1K(b *testing.B)   { benchSitemap(1000, b) }
func BenchmarkSitemap10K(b *testing.B)  { benchSitemap(10000, b) }
func BenchmarkSitemap100K(b *testing.B) { benchSitemap(100000, b) }
func BenchmarkSitemap1M(b *testing.B)   { benchSitemap(1000000, b) }
