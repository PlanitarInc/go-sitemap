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
	//     <loc>http://goiguide.com/</loc>
	//     <lastmod>2025-11-02T11:34:58Z</lastmod>
	//   </url>
	//   <url>
	//     <loc>http://goiguide.com/test/</loc>
	//     <image:image>
	//       <image:loc>http://goiguide.com/test/1.jpg</image:loc>
	//     </image:image>
	//     <image:image>
	//       <image:loc>http://goiguide.com/test/2.jpg</image:loc>
	//     </image:image>
	//     <image:image>
	//       <image:loc>http://goiguide.com/test/3.jpg</image:loc>
	//     </image:image>
	//   </url>
	// </urlset>
	//
	// ::: Index
	//
	// <?xml version="1.0" encoding="UTF-8"?>
	// <sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	//   <sitemap>
	//     <loc>https://goiguide.com/sitemap-0.xml</loc>
	//   </sitemap>
	// </sitemapindex>
}

func ExampleWriteAll_buffers() {
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
	//     <loc>http://goiguide.com/</loc>
	//     <lastmod>2025-11-02T11:34:58Z</lastmod>
	//   </url>
	//   <url>
	//     <loc>http://goiguide.com/test/</loc>
	//     <image:image>
	//       <image:loc>http://goiguide.com/test/1.jpg</image:loc>
	//     </image:image>
	//     <image:image>
	//       <image:loc>http://goiguide.com/test/2.jpg</image:loc>
	//     </image:image>
	//     <image:image>
	//       <image:loc>http://goiguide.com/test/3.jpg</image:loc>
	//     </image:image>
	//   </url>
	// </urlset>
	//
	// ::: Index
	//
	// <?xml version="1.0" encoding="UTF-8"?>
	// <sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	//   <sitemap>
	//     <loc>https://goiguide.com/sitemap-0.xml</loc>
	//   </sitemap>
	// </sitemapindex>
}

func ExampleWriteAll_dynamicInput() {
	err := sitemap.WriteAll(&stdoutOutput{}, &dynamicInput{Length: 3})
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	// Output:
	// ::: Urlset 0

	// <?xml version="1.0" encoding="UTF-8"?>
	// <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
	//   <url>
	//     <loc>https://goiguide.com/entry-00000</loc>
	//     <lastmod>2020-10-31T11:00:00Z</lastmod>
	//     <image:image>
	//       <image:loc>https://goiguide.com/entry-00000/thumb.png</image:loc>
	//     </image:image>
	//   </url>
	//   <url>
	//     <loc>https://goiguide.com/entry-00001</loc>
	//     <lastmod>2020-11-01T11:00:00Z</lastmod>
	//     <image:image>
	//       <image:loc>https://goiguide.com/entry-00001/thumb.png</image:loc>
	//     </image:image>
	//   </url>
	//   <url>
	//     <loc>https://goiguide.com/entry-00002</loc>
	//     <lastmod>2020-11-02T11:00:00Z</lastmod>
	//     <image:image>
	//       <image:loc>https://goiguide.com/entry-00002/thumb.png</image:loc>
	//     </image:image>
	//   </url>
	// </urlset>

	// ::: Index

	// <?xml version="1.0" encoding="UTF-8"?>
	// <sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	//   <sitemap>
	//     <loc>https://goiguide.com/sitemap-00.xml</loc>
	//   </sitemap>
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
	Arr     []sitemap.UrlEntry
	nextIdx int
}

func (a *arrayInput) Next() *sitemap.UrlEntry {
	if a.nextIdx >= len(a.Arr) {
		return nil
	}

	a.nextIdx++
	return &a.Arr[a.nextIdx-1]
}

func (a *arrayInput) GetUrlsetUrl(n int) string {
	return fmt.Sprintf("https://goiguide.com/sitemap-%d.xml", n)
}

type dynamicInput struct {
	Length  int
	nextIdx int
	entry   sitemap.UrlEntry
}

func (d *dynamicInput) Next() *sitemap.UrlEntry {
	if d.nextIdx >= d.Length {
		return nil
	}

	idx := d.nextIdx
	d.nextIdx++
	d.entry.Loc = fmt.Sprintf("https://goiguide.com/entry-%05d", idx)
	d.entry.LastMod = time.Date(2020, time.November, idx, 11, 0, 0, 0, time.UTC)
	d.entry.Images = []string{fmt.Sprintf("https://goiguide.com/entry-%05d/thumb.png", idx)}
	return &d.entry
}

func (d dynamicInput) GetUrlsetUrl(n int) string {
	return fmt.Sprintf("https://goiguide.com/sitemap-%02d.xml", n)
}
