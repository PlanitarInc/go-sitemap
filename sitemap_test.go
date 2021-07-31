package sitemap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestWriteAll(t *testing.T) {
	customEntry := func(idx int) UrlEntry {
		return simpleEntry{
			Loc:     fmt.Sprintf("http://goiguide.com/%d", idx),
			LastMod: minDate.AddDate(1, 2, 3),
			Images: []string{
				fmt.Sprintf("http://youriguide.com/%d.jpg", idx),
			},
		}
	}
	customUrl := func(idx int) string {
		return fmt.Sprintf("urlset %03d", idx)
	}

	type simpleSitemap struct {
		Locs []string `xml:"url>loc"`
	}

	assertOutput := func(out *bufferOuput, in *dynamicInput) {
		nsitemaps := int(math.Ceil(float64(in.Size) / 50_000))

		var index simpleSitemap
		Ω(xml.Unmarshal(out.index.Bytes(), &index)).Should(BeNil())
		Ω(index.Locs).Should(HaveLen(nsitemaps))
		for i := 0; i < nsitemaps; i++ {
			Ω(index.Locs[i]).Should(Equal(fmt.Sprintf("urlset %03d", i)))
		}

		Ω(out.sitemaps).Should(HaveLen(nsitemaps))
		for i := 0; i < nsitemaps; i++ {
			var s simpleSitemap
			Ω(xml.Unmarshal(out.sitemaps[i].Bytes(), &s)).Should(BeNil())

			urlsetOffset := i * 50_000
			nlocs := in.Size - urlsetOffset
			if nlocs > 50_000 {
				nlocs = 50_000
			}
			Ω(s.Locs).Should(HaveLen(nlocs))

			for j := 0; j < nlocs; j++ {
				Ω(s.Locs[j]).Should(Equal(fmt.Sprintf("http://goiguide.com/%d", urlsetOffset+j)))
			}
		}
	}

	t.Run("empty", func(t *testing.T) {
		RegisterTestingT(t)

		in := dynamicInput{
			CustomEntry:     customEntry,
			CustomUrlsetUrl: customUrl,
		}
		var out bufferOuput

		Ω(WriteAll(&out, &in)).Should(BeNil())
		Ω(out.index.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>urlset 000</loc>
  </url>
</sitemapindex>
		`)))

		Ω(out.sitemaps).Should(HaveLen(1))
		Ω(out.sitemaps[0].String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
</urlset>
		`)))
	})

	t.Run("shortSingleSitemap", func(t *testing.T) {
		RegisterTestingT(t)

		in := dynamicInput{
			Size:            3,
			CustomEntry:     customEntry,
			CustomUrlsetUrl: customUrl,
		}
		var out bufferOuput

		Ω(WriteAll(&out, &in)).Should(BeNil())
		Ω(out.index.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>urlset 000</loc>
  </url>
</sitemapindex>
		`)))

		Ω(out.sitemaps).Should(HaveLen(1))
		Ω(out.sitemaps[0].String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
  <url>
    <loc>http://goiguide.com/0</loc>
    <lastmod>2001-03-04T00:00:00Z</lastmod>
    <image:image>
      <image:loc>http://youriguide.com/0.jpg</image:loc>
    </image:image>
  </url>
  <url>
    <loc>http://goiguide.com/1</loc>
    <lastmod>2001-03-04T00:00:00Z</lastmod>
    <image:image>
      <image:loc>http://youriguide.com/1.jpg</image:loc>
    </image:image>
  </url>
  <url>
    <loc>http://goiguide.com/2</loc>
    <lastmod>2001-03-04T00:00:00Z</lastmod>
    <image:image>
      <image:loc>http://youriguide.com/2.jpg</image:loc>
    </image:image>
  </url>
</urlset>
		`)))
	})

	t.Run("maxSingleSitemap", func(t *testing.T) {
		RegisterTestingT(t)

		in := dynamicInput{
			Size:            50_000,
			CustomEntry:     customEntry,
			CustomUrlsetUrl: customUrl,
		}
		var out bufferOuput

		Ω(WriteAll(&out, &in)).Should(BeNil())
		Ω(out.index.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>urlset 000</loc>
  </url>
</sitemapindex>
		`)))

		assertOutput(&out, &in)
	})

	t.Run("minTwoSitemaps", func(t *testing.T) {
		RegisterTestingT(t)

		in := dynamicInput{
			Size:            50_000 + 1,
			CustomEntry:     customEntry,
			CustomUrlsetUrl: customUrl,
		}
		var out bufferOuput

		Ω(WriteAll(&out, &in)).Should(BeNil())
		Ω(out.index.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>urlset 000</loc>
  </url>
  <url>
    <loc>urlset 001</loc>
  </url>
</sitemapindex>
		`)))

		assertOutput(&out, &in)
	})

	t.Run("maxTwoSitemaps", func(t *testing.T) {
		RegisterTestingT(t)

		in := dynamicInput{
			Size:            50_000 * 2,
			CustomEntry:     customEntry,
			CustomUrlsetUrl: customUrl,
		}
		var out bufferOuput

		Ω(WriteAll(&out, &in)).Should(BeNil())
		assertOutput(&out, &in)
	})

	t.Run("multipleSitemaps", func(t *testing.T) {
		RegisterTestingT(t)

		in := dynamicInput{
			Size:            50_000*3 + 123,
			CustomEntry:     customEntry,
			CustomUrlsetUrl: customUrl,
		}
		var out bufferOuput

		Ω(WriteAll(&out, &in)).Should(BeNil())
		assertOutput(&out, &in)
	})

	t.Run("numerousSitemaps", func(t *testing.T) {
		RegisterTestingT(t)

		in := dynamicInput{
			Size:            50_000*11 + 97,
			CustomEntry:     customEntry,
			CustomUrlsetUrl: customUrl,
		}
		var out bufferOuput

		Ω(WriteAll(&out, &in)).Should(BeNil())
		assertOutput(&out, &in)
	})

	t.Run("failures", func(t *testing.T) {
		t.Run("urlset", func(t *testing.T) {
			RegisterTestingT(t)

			in := dynamicInput{
				CustomEntry:     customEntry,
				CustomUrlsetUrl: customUrl,
			}
			out := failiingOutput{
				FailUrlset: true,
			}

			Ω(WriteAll(&out, &in)).Should(MatchError("failingWriter error"))
		})

		t.Run("index", func(t *testing.T) {
			RegisterTestingT(t)

			in := dynamicInput{
				CustomEntry:     customEntry,
				CustomUrlsetUrl: customUrl,
			}
			out := failiingOutput{
				FailIndex: true,
			}

			Ω(WriteAll(&out, &in)).Should(MatchError("failingWriter error"))
		})

		t.Run("both", func(t *testing.T) {
			RegisterTestingT(t)

			in := dynamicInput{
				CustomEntry:     customEntry,
				CustomUrlsetUrl: customUrl,
			}
			out := failiingOutput{
				FailUrlset: true,
				FailIndex:  true,
			}

			Ω(WriteAll(&out, &in)).Should(MatchError("failingWriter error"))
		})
	})
}

func TestSitemapWriter_WriteUrlsetFile(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		RegisterTestingT(t)

		var s sitemapWriter
		var out bytes.Buffer
		Ω(s.writeUrlsetFile(&out, &arrayInput{})).Should(BeNil())
		Ω(out.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
</urlset>
		`)))

		out.Reset()
		Ω(s.writeUrlsetFile(&out, &arrayInput{Arr: []simpleEntry{{}, {}, {}, {}}})).Should(BeNil())
		Ω(out.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
  <url>
    <loc></loc>
  </url>
  <url>
    <loc></loc>
  </url>
  <url>
    <loc></loc>
  </url>
  <url>
    <loc></loc>
  </url>
</urlset>
		`)))
	})

	t.Run("simple", func(t *testing.T) {
		RegisterTestingT(t)

		entries := []simpleEntry{
			{
				Loc:     "one",
				LastMod: time.Date(1999, 12, 31, 23, 59, 59, 0, time.UTC),
			},
			{
				Loc:     "two",
				LastMod: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				Loc:     "three",
				LastMod: time.Date(2015, 7, 22, 15, 48, 2, 0, time.UTC),
			},
		}

		var s sitemapWriter
		var out bytes.Buffer
		Ω(s.writeUrlsetFile(&out, &arrayInput{Arr: entries})).Should(BeNil())
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
	})

	t.Run("images", func(t *testing.T) {
		RegisterTestingT(t)

		entries := []simpleEntry{
			{
				Images: []string{},
			},
			{},
		}

		var s sitemapWriter
		var out bytes.Buffer
		Ω(s.writeUrlsetFile(&out, &arrayInput{Arr: entries})).Should(BeNil())
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

		entries = []simpleEntry{
			{
				Loc:    "one",
				Images: []string{"a", "b", "c"},
			},
			{
				Loc: "two",
			},
			{
				Loc:    "three",
				Images: []string{"w", "x", "y", "z"},
			},
		}

		out.Reset()
		Ω(s.writeUrlsetFile(&out, &arrayInput{Arr: entries})).Should(BeNil())
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
	})

	t.Run("escaping", func(t *testing.T) {
		RegisterTestingT(t)

		entries := []simpleEntry{
			{
				Loc:    `http://www.example.com/q="<'a'&'b'>"`,
				Images: []string{`"<`, `qwe&qw&ewq`, `asd`},
			},
		}

		var s sitemapWriter
		var out bytes.Buffer
		Ω(s.writeUrlsetFile(&out, &arrayInput{Arr: entries})).Should(BeNil())
		Ω(out.String()).Should(Equal(strings.TrimSpace(`
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
	})

	t.Run("failures", func(t *testing.T) {
		t.Run("errMaxCapReached", func(t *testing.T) {
			RegisterTestingT(t)

			in := dynamicInput{
				DefaultEntry: simpleEntry{
					Loc:     "http://www.example.com/qweqwe",
					LastMod: minDate.AddDate(1, 2, 3),
					Images:  []string{"http://www.example.com/qweqwe/thumb.jpg"},
				},
				Size: 50_000 + 1,
			}

			var s sitemapWriter
			Ω(s.writeUrlsetFile(ioutil.Discard, &in)).
				Should(MatchError("max 50K capacity is reached"))
		})

		t.Run("failingWriter", func(t *testing.T) {
			RegisterTestingT(t)

			in := dynamicInput{
				DefaultEntry: simpleEntry{
					Loc:     "http://www.example.com/qweqwe",
					LastMod: minDate.AddDate(1, 2, 3),
					Images:  []string{"http://www.example.com/qweqwe/thumb.jpg"},
				},
				Size: 50_000 + 1,
			}

			var s sitemapWriter
			Ω(s.writeUrlsetFile(&failingWriter{}, &in)).
				Should(MatchError("failingWriter error"))
		})
	})
}

func TestSitemapWriter_WriteIndexFile(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		RegisterTestingT(t)

		emptyUrl := func(idx int) string {
			return ""
		}

		var s sitemapWriter
		var out bytes.Buffer
		Ω(s.writeIndexFile(&out, &arrayInput{CustomUrlsetUrl: emptyUrl}, 0)).Should(BeNil())
		Ω(out.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
</sitemapindex>
		`)))
	})

	t.Run("default", func(t *testing.T) {
		RegisterTestingT(t)

		var s sitemapWriter
		var out bytes.Buffer
		Ω(s.writeIndexFile(&out, &arrayInput{}, 3)).Should(BeNil())
		Ω(out.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>urlset no. 1</loc>
  </url>
  <url>
    <loc>urlset no. 2</loc>
  </url>
  <url>
    <loc>urlset no. 3</loc>
  </url>
</sitemapindex>
		`)))
	})

	t.Run("custom", func(t *testing.T) {
		RegisterTestingT(t)

		simpleUrl := func(idx int) string {
			return fmt.Sprintf("custom @%03d@", idx)
		}

		var s sitemapWriter
		var out bytes.Buffer
		Ω(s.writeIndexFile(&out, &arrayInput{CustomUrlsetUrl: simpleUrl}, 4)).Should(BeNil())
		Ω(out.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>custom @000@</loc>
  </url>
  <url>
    <loc>custom @001@</loc>
  </url>
  <url>
    <loc>custom @002@</loc>
  </url>
  <url>
    <loc>custom @003@</loc>
  </url>
</sitemapindex>
		`)))
	})

	t.Run("escaping", func(t *testing.T) {
		RegisterTestingT(t)

		fancyUrl := func(idx int) string {
			switch idx {
			case 0:
				return `http://www.example.com/q="<'a'&'b'>`
			case 1:
				return "🥴.com/"
			case 2:
				return "гоуайгайд.ком/"
			default:
				return fmt.Sprintf("🤟.🤙/?idx=<%02d>&e=/'🤪?", idx)
			}
		}

		var s sitemapWriter
		var out bytes.Buffer
		Ω(s.writeIndexFile(&out, &arrayInput{CustomUrlsetUrl: fancyUrl}, 5)).Should(BeNil())
		Ω(out.String()).Should(Equal(strings.TrimSpace(`
<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>http://www.example.com/q=&#34;&lt;&#39;a&#39;&amp;&#39;b&#39;&gt;</loc>
  </url>
  <url>
    <loc>🥴.com/</loc>
  </url>
  <url>
    <loc>гоуайгайд.ком/</loc>
  </url>
  <url>
    <loc>🤟.🤙/?idx=&lt;03&gt;&amp;e=/&#39;🤪?</loc>
  </url>
  <url>
    <loc>🤟.🤙/?idx=&lt;04&gt;&amp;e=/&#39;🤪?</loc>
  </url>
</sitemapindex>
		`)))
	})

	t.Run("failures", func(t *testing.T) {
		t.Run("failingWriter", func(t *testing.T) {
			RegisterTestingT(t)

			in := dynamicInput{
				CustomUrlsetUrl: func(idx int) string {
					return fmt.Sprintf("urlset %03d", idx)
				},
			}

			var s sitemapWriter
			Ω(s.writeIndexFile(&failingWriter{}, &in, 100)).
				Should(MatchError("failingWriter error"))
		})
	})
}

type arrayInput struct {
	Arr             []simpleEntry
	CustomUrlsetUrl func(int) string

	nextIdx int
}

func (a arrayInput) HasNext() bool {
	return a.nextIdx < len(a.Arr)
}

func (a *arrayInput) Next() UrlEntry {
	idx := a.nextIdx
	a.nextIdx++
	return a.Arr[idx]
}

func (a *arrayInput) GetUrlsetUrl(idx int) string {
	if a.CustomUrlsetUrl != nil {
		return a.CustomUrlsetUrl(idx)
	}

	return fmt.Sprintf("urlset no. %d", idx+1)
}

type simpleEntry struct {
	Loc     string
	LastMod time.Time
	Images  []string
}

func (e simpleEntry) GetLoc() string {
	return e.Loc
}

func (e simpleEntry) GetLastMod() time.Time {
	return e.LastMod
}

func (e simpleEntry) GetImages() []string {
	return e.Images
}

type failiingOutput struct {
	FailIndex  bool
	FailUrlset bool
}

func (o *failiingOutput) Index() io.Writer {
	if o.FailIndex {
		return failingWriter{}
	}

	return io.Discard
}

func (o *failiingOutput) Urlset() io.Writer {
	if o.FailUrlset {
		return failingWriter{}
	}

	return io.Discard
}

type failingWriter struct{}

func (failingWriter) Write(bs []byte) (int, error) {
	return 0, errors.New("failingWriter error")
}

type bufferOuput struct {
	index    bytes.Buffer
	sitemaps []bytes.Buffer
}

func (o *bufferOuput) Index() io.Writer {
	return &o.index
}

func (o *bufferOuput) Urlset() io.Writer {
	o.sitemaps = append(o.sitemaps, bytes.Buffer{})
	return &o.sitemaps[len(o.sitemaps)-1]
}

type dynamicInput struct {
	Size            int
	DefaultEntry    simpleEntry
	CustomEntry     func(int) UrlEntry
	CustomUrlsetUrl func(int) string

	nextIdx int
}

func (d *dynamicInput) Reset() {
	d.nextIdx = 0
}

func (d dynamicInput) HasNext() bool {
	return d.nextIdx < d.Size
}

func (d *dynamicInput) Next() UrlEntry {
	idx := d.nextIdx
	d.nextIdx++
	if d.CustomEntry != nil {
		return d.CustomEntry(idx)
	}

	return d.DefaultEntry
}

func (d *dynamicInput) GetUrlsetUrl(idx int) string {
	if d.CustomUrlsetUrl != nil {
		return d.CustomUrlsetUrl(idx)
	}

	return ""
}

type discardOutput struct{}

func (discardOutput) Index() io.Writer {
	return ioutil.Discard
}

func (discardOutput) Urlset() io.Writer {
	return ioutil.Discard
}

func BenchmarkWriteAll(b *testing.B) {
	var out discardOutput
	in := dynamicInput{
		DefaultEntry: simpleEntry{
			Loc:     "http://www.example.com/qweqwe",
			LastMod: minDate.AddDate(1, 2, 3),
			Images:  []string{"http://www.example.com/qweqwe/thumb.jpg"},
		},
		CustomUrlsetUrl: func(int) string {
			return "const url"
		},
	}

	for p := 0; p < 7; p++ {
		size := int(math.Pow10(p))
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			in.Size = size
			for n := 0; n < b.N; n++ {
				in.Reset()
				_ = WriteAll(out, &in)
			}
		})
	}
}

func BenchmarkWriteUrlset(b *testing.B) {
	in := dynamicInput{
		DefaultEntry: simpleEntry{
			Loc:     "http://www.example.com/qweqwe",
			LastMod: minDate.AddDate(1, 2, 3),
			Images:  []string{"http://www.example.com/qweqwe/thumb.jpg"},
		},
		CustomUrlsetUrl: func(int) string {
			return "const url"
		},
	}
	var s sitemapWriter

	for p := 0; p < 7; p++ {
		size := int(math.Pow10(p))
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			in.Size = size
			for n := 0; n < b.N; n++ {
				in.Reset()
				_ = s.writeUrlsetFile(ioutil.Discard, &in)
			}
		})
	}
}

func BenchmarkWriteIndex(b *testing.B) {
	in := dynamicInput{
		DefaultEntry: simpleEntry{
			Loc:     "http://www.example.com/qweqwe",
			LastMod: minDate.AddDate(1, 2, 3),
			Images:  []string{"http://www.example.com/qweqwe/thumb.jpg"},
		},
		CustomUrlsetUrl: func(int) string {
			return "const url"
		},
	}
	var s sitemapWriter

	for p := 0; p < 6; p++ {
		nfiles := int(math.Pow10(p))
		b.Run(strconv.Itoa(nfiles), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				in.Reset()
				_ = s.writeIndexFile(ioutil.Discard, &in, nfiles)
			}
		})
	}
}

func TestSitemapWriter_writeXmlString(t *testing.T) {
	RegisterTestingT(t)

	testCases := []struct {
		In, Out string
	}{
		{},
		{In: "abc", Out: "abc"},
		{In: "<b>", Out: "&lt;b&gt;"},
		{In: "c;d", Out: "c;d"},
		{In: "&", Out: "&amp;"},
		{In: `'="`, Out: "&#39;=&#34;"},
		{
			In:  `a'b"c&d<e>f` + "\tg\nh\ri",
			Out: "a&#39;b&#34;c&amp;d&lt;e&gt;f&#x9;g&#xA;h&#xD;i",
		},
		{
			In:  `https://goiguide.com/showcase`,
			Out: `https://goiguide.com/showcase`,
		},
		{
			In:  `https://goiguide.com/showcase?a=1&b=2`,
			Out: `https://goiguide.com/showcase?a=1&amp;b=2`,
		},
	}

	var s sitemapWriter
	for _, tc := range testCases {
		var b bytes.Buffer
		s.writeXmlString(&b, tc.In)
		Ω(b.String()).Should(Equal(tc.Out), tc.In)
	}
}

func BenchmarkSitemapWriter_writeXmlString(b *testing.B) {
	b.Run("short", func(b *testing.B) {
		var s sitemapWriter
		for n := 0; n < b.N; n++ {
			s.writeXmlString(ioutil.Discard, "short")
		}
	})

	b.Run("long", func(b *testing.B) {
		var s sitemapWriter
		for n := 0; n < b.N; n++ {
			s.writeXmlString(ioutil.Discard,
				"https://goiguide.com/showcase?a=1&b=2")
		}
	})
}

func TestSitemapWriter_writeXmlTime(t *testing.T) {
	RegisterTestingT(t)

	tz1 := time.FixedZone("negative", -15*60*60)
	tz2 := time.FixedZone("positive", 15*60*60)

	testCases := []struct {
		In  time.Time
		Out string
	}{
		{
			Out: "0001-01-01T00:00:00Z",
		},
		{
			In:  time.Date(1999, time.December, 31, 23, 59, 59, 0, time.UTC),
			Out: "1999-12-31T23:59:59Z",
		},
		{
			In:  time.Date(2020, time.March, 15, 12, 13, 14, 999, time.UTC),
			Out: "2020-03-15T12:13:14Z",
		},
		{
			In:  time.Date(2021, time.July, 31, 23, 59, 59, 7, tz1),
			Out: "2021-07-31T23:59:59-15:00",
		},
		{
			In:  time.Date(2022, time.November, 29, 23, 59, 59, 7, tz2),
			Out: "2022-11-29T23:59:59+15:00",
		},
	}

	var s sitemapWriter
	for _, tc := range testCases {
		var b bytes.Buffer
		s.writeXmlTime(&b, tc.In)
		Ω(b.String()).Should(Equal(tc.Out), tc.In.Format(time.RFC3339))
	}
}

func BenchmarkSitemapWriter_writeXmlTime(b *testing.B) {
	b.Run("utc", func(b *testing.B) {
		var s sitemapWriter
		for n := 0; n < b.N; n++ {
			t := time.Date(2020, time.March, n, 12, 13, 14, 7, time.UTC)
			s.writeXmlTime(ioutil.Discard, t)
		}
	})

	b.Run("custom", func(b *testing.B) {
		tz1 := time.FixedZone("negative", -15*60*60)
		var s sitemapWriter
		for n := 0; n < b.N; n++ {
			t := time.Date(2020, time.March, n, 12, 13, 14, 7, tz1)
			s.writeXmlTime(ioutil.Discard, t)
		}
	})
}
