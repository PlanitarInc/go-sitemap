package sitemap

import (
	"bytes"
	"encoding/xml"
	"io"
	"time"
)

// WriteAll writes all files to the given output. Urlset files are written to
// writers provided by o.Urlset(), the function will call it every time a new
// file is to be written. The final index file is written to a writer provided
// by o.Index().
// The function aborts if any unexpected error occurs when writing.
func WriteAll(o Output, in Input) error {
	var s sitemapWriter
	var nfiles int
	var carryOverEntry *UrlEntry
	for {
		nfiles++
		var err error
		carryOverEntry, err = s.writeUrlsetFile(o.Urlset(), in, carryOverEntry)
		if err != nil {
			return err
		}

		if carryOverEntry == nil {
			return s.writeIndexFile(o.Index(), in, nfiles)
		}
	}
}

type sitemapWriter struct {
	// temporary buffer used to escape string values for XML
	buf bytes.Buffer
}

// writeIndexFile writes Sitemap index file for N files.
func (s *sitemapWriter) writeIndexFile(w io.Writer, in Input, nfiles int) error {
	abortWriter := abortWriter{underlying: w}

	_, _ = abortWriter.Write(indexHeader)
	for i := 0; i < nfiles; i++ {
		s.writeXmlSitemapLoc(&abortWriter, in.GetUrlsetUrl(i))
	}
	_, _ = abortWriter.Write(indexFooter)

	return abortWriter.firstErr
}

// writeUrlsetFile writes a single Sitemap Urlset file for the first 50K entries
// in the given input.
func (s *sitemapWriter) writeUrlsetFile(
	w io.Writer,
	in Input,
	prevEntry *UrlEntry,
) (*UrlEntry, error) {
	abortWriter := abortWriter{underlying: w}

	_, _ = abortWriter.Write(urlsetHeader)

	var count int
	// This is a continuation of a previous iteration. Write the carry-over
	// entry without calling "Next()". Otherwise, we would lose an entry.
	if prevEntry != nil {
		s.writeXmlUrlEntry(&abortWriter, prevEntry)
		count++
	}

	var carryOverEntry *UrlEntry
	for ; ; count++ {
		entry := in.Next()
		if entry == nil {
			break
		}

		if count >= maxSitemapCap {
			carryOverEntry = entry
			break
		}

		s.writeXmlUrlEntry(&abortWriter, entry)
	}
	_, _ = abortWriter.Write(urlsetFooter)

	if abortWriter.firstErr != nil {
		return nil, abortWriter.firstErr
	}

	return carryOverEntry, nil
}

func (s *sitemapWriter) writeXmlUrlEntry(w io.Writer, e *UrlEntry) {
	_, _ = w.Write(tagUrlOpen)
	_, _ = w.Write(tagLocOpen)
	s.writeXmlString(w, e.Loc)
	_, _ = w.Write(tagLocClose)
	if !e.LastMod.Before(minDate) {
		_, _ = w.Write(tagLastmodOpen)
		s.writeXmlTime(w, e.LastMod)
		_, _ = w.Write(tagLastmodClose)
	}
	if len(e.Images) > 0 {
		for i := range e.Images {
			_, _ = w.Write(tagImageOpen)
			s.writeXmlString(w, e.Images[i])
			_, _ = w.Write(tagImageClose)
		}
	}
	_, _ = w.Write(tagUrlClose)
}

func (s *sitemapWriter) writeXmlSitemapLoc(w io.Writer, loc string) {
	_, _ = w.Write(tagSitemapOpen)
	_, _ = w.Write(tagLocOpen)
	s.writeXmlString(w, loc)
	_, _ = w.Write(tagLocClose)
	_, _ = w.Write(tagSitemapClose)
}

func (s *sitemapWriter) writeXmlString(w io.Writer, str string) {
	// Here we try to perform an "alloc-free" conversion of a dynamic string
	// to a byte slice using a temporary buffer.
	s.buf.Reset()
	_, _ = s.buf.WriteString(str)
	_ = xml.EscapeText(w, s.buf.Bytes())
}

func (s *sitemapWriter) writeXmlTime(w io.Writer, t time.Time) {
	// Here we try to perform an "alloc-free" conversion of a dynamic date
	// to a byte slice using a temporary buffer.
	s.buf.Reset()
	s.buf.Grow(len(time.RFC3339) * 2) // *2 is just in case
	bs := t.AppendFormat(s.buf.Bytes(), time.RFC3339)
	_, _ = w.Write(bs)
}

// Below are constant strings converted to byte slices ahead of time
// to avoid run-time allocations caused by string to byte slice conversions.
var (
	indexHeader = []byte(xml.Header +
		`<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` +
		"\n",
	)
	indexFooter = []byte("</sitemapindex>")

	urlsetHeader = []byte(xml.Header +
		`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">` +
		"\n",
	)
	urlsetFooter = []byte(`</urlset>`)

	tagSitemapOpen  = []byte("  <sitemap>\n")
	tagSitemapClose = []byte("  </sitemap>\n")
	tagUrlOpen      = []byte("  <url>\n")
	tagUrlClose     = []byte("  </url>\n")
	tagLocOpen      = []byte("    <loc>")
	tagLocClose     = []byte("</loc>\n")
	tagLastmodOpen  = []byte("    <lastmod>")
	tagLastmodClose = []byte("</lastmod>\n")
	tagImageOpen    = []byte("    <image:image>\n      <image:loc>")
	tagImageClose   = []byte("</image:loc>\n    </image:image>\n")
)

var minDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

type abortWriter struct {
	underlying io.Writer
	firstErr   error
}

func (w *abortWriter) Write(p []byte) (n int, err error) {
	if w.firstErr != nil {
		return 0, w.firstErr
	}

	n, err = w.underlying.Write(p)
	if err != nil {
		w.firstErr = err
	}
	return
}

const (
	maxSitemapCap = 50_000
)
