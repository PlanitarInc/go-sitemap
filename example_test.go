package sitemap_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/PlanitarInc/go-sitemap"
)

func ExampleWriteAll_stdout() {
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

	err := sitemap.WriteAll(&stdoutOutput{}, &arrayInput{Arr: entries})
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	// Output:
	// ::: Urlset 0
	//
	// <?xml version="1.0" encoding="UTF-8"?>
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
	//
	// ::: Index
	//
	// <?xml version="1.0" encoding="UTF-8"?>
	// <sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	//   <url>
	//     <loc>https://example.com/sitemap-0.xml</loc>
	//   </url>
	// </sitemapindex>
}

func ExampleWriteAll_buffers() {
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

	var out bufferOutput
	err := sitemap.WriteAll(&out, &arrayInput{Arr: entries})
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	for i := range out.UrlsetBufs {
		fmt.Printf("\n\n::: Urlset %d\n\n", i)
		fmt.Print(out.UrlsetBufs[i].String())
	}
	fmt.Printf("\n\n::: Index\n\n")
	fmt.Print(out.IndexBuf.String())

	// Output:
	// ::: Urlset 0
	//
	// <?xml version="1.0" encoding="UTF-8"?>
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
	//
	// ::: Index
	//
	// <?xml version="1.0" encoding="UTF-8"?>
	// <sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	//   <url>
	//     <loc>https://example.com/sitemap-0.xml</loc>
	//   </url>
	// </sitemapindex>
}

type stdoutOutput struct {
	urlsetIdx int
}

func (o stdoutOutput) Index() io.Writer {
	fmt.Fprintf(os.Stdout, "\n\n::: Index\n\n")
	return os.Stdout
}

func (o *stdoutOutput) Urlset() io.Writer {
	fmt.Fprintf(os.Stdout, "\n\n::: Urlset %d\n\n", o.urlsetIdx)
	o.urlsetIdx++
	return os.Stdout
}

type bufferOutput struct {
	IndexBuf   bytes.Buffer
	UrlsetBufs []bytes.Buffer
}

func (o *bufferOutput) Index() io.Writer {
	return &o.IndexBuf
}

func (o *bufferOutput) Urlset() io.Writer {
	o.UrlsetBufs = append(o.UrlsetBufs, bytes.Buffer{})
	return &o.UrlsetBufs[len(o.UrlsetBufs)-1]
}

type arrayInput struct {
	Arr     []SimpleEntry
	NextIdx int
}

func (a arrayInput) HasNext() bool {
	return a.NextIdx < len(a.Arr)
}

func (a *arrayInput) Next() sitemap.UrlEntry {
	idx := a.NextIdx
	a.NextIdx++
	return a.Arr[idx]
}

func (a *arrayInput) GetUrlsetUrl(n int) string {
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
