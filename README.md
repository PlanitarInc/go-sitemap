[![Go Reference](https://pkg.go.dev/badge/github.com/PlanitarInc/go-sitemap.svg)](https://pkg.go.dev/github.com/PlanitarInc/go-sitemap)
![CI Status](https://github.com/PlanitarInc/go-sitemap/actions/workflows/ci-flow.yml/badge.svg?branch=master) [![Coverage Status](https://coveralls.io/repos/github/PlanitarInc/go-sitemap/badge.svg)](https://coveralls.io/github/PlanitarInc/go-sitemap)

A GO library for generating [Sitemap XML files](https://www.sitemaps.org/protocol.html).

### Example

<!-- GO Playground: https://play.golang.org/p/PTYJXlJt8ep -->

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
	Arr     []sitemap.UrlEntry
	nextIdx int
}

func (a *ArrayInput) Next() *sitemap.UrlEntry {
	if a.nextIdx >= len(a.Arr) {
		return nil
	}

	a.nextIdx++
	return &a.Arr[a.nextIdx-1]
}

func (a *ArrayInput) GetUrlsetUrl(n int) string {
	return fmt.Sprintf("https://goiguide.com/sitemap-%d.xml", n)
}

func main() {
	entries := []sitemap.UrlEntry{
		{
			Loc:     "http://goiguide.com/",
			LastMod: time.Date(2025, time.November, 2, 11, 34, 58, 123, time.UTC),
		},
		{
			Loc: "http://goiguide.com/test/",
			Images: []string{
				"http://goiguide.com/test/1.jpg",
				"http://goiguide.com/test/2.jpg",
				"http://goiguide.com/test/3.jpg",
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
    <loc>http://goiguide.com/</loc>
    <lastmod>2025-11-02T11:34:58Z</lastmod>
  </url>
  <url>
    <loc>http://goiguide.com/test/</loc>
    <image:image>
      <image:loc>http://goiguide.com/test/1.jpg</image:loc>
    </image:image>
    <image:image>
      <image:loc>http://goiguide.com/test/2.jpg</image:loc>
    </image:image>
    <image:image>
      <image:loc>http://goiguide.com/test/3.jpg</image:loc>
    </image:image>
  </url>
</urlset>

::: Index

<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap>
    <loc>https://goiguide.com/sitemap-0.xml</loc>
  </sitemap>
</sitemapindex>
```

### Benchmark

The library is trying to be smart when producing XML and keeps the allocations
constant. That is, space complexity is `O(1)` -- it does not depend on the
number of entries.

```
$ go test -run '^$' -bench . -benchmem
goos: darwin
goarch: amd64
pkg: github.com/PlanitarInc/go-sitemap
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkWriteAll/1-12                      	 1328336	       945.2 ns/op	     128 B/op	       3 allocs/op
BenchmarkWriteAll/10-12                     	  150031	      6808 ns/op	     128 B/op	       3 allocs/op
BenchmarkWriteAll/100-12                    	   17624	     62340 ns/op	     128 B/op	       3 allocs/op
BenchmarkWriteAll/1000-12                   	    1898	    627547 ns/op	     128 B/op	       3 allocs/op
BenchmarkWriteAll/10000-12                  	     192	   6315018 ns/op	     128 B/op	       3 allocs/op
BenchmarkWriteAll/100000-12                 	      18	  62742897 ns/op	     160 B/op	       4 allocs/op
BenchmarkWriteAll/1000000-12                	       2	 632302752 ns/op	     736 B/op	      22 allocs/op
```
