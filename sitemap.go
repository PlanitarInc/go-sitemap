package sitemap

import (
	"encoding/xml"
	"errors"
	"io"
	"time"
)

type Url struct {
	XMLName xml.Name `xml:"url"`
	Loc     string   `xml:"loc"`
	LastMod string   `xml:"lastmod,omitempty"`
	Images  []Image  `xml:"image:image,omitempty"`
	//      Priority   float32   `xml:"priority"`
	//      ChangeFreq string    `xml:"changefreq,omitempty"`
}

type Image struct {
	Loc string `xml:"image:loc"`
}

var minDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

func t2str(t time.Time) string {
	if t.Before(minDate) {
		return ""
	}

	return t.Format(time.RFC3339)
}

// WriteAll writes all files to the given output. Urlset files are written to
// writers provided by o.Urlset(), the function will call it every time a new
// file is to be written. The final index file is written to a writer provided
// by o.Index().
// The function aborts if any unexpected error occurs when writing.
func WriteAll(o Output, in Input) error {
	var nfiles int
	for {
		nfiles++
		err := writeUrlsetFile(o.Urlset(), in)
		if err != nil && !errors.Is(err, errMaxCapReached{}) {
			return err
		}

		if err == nil {
			return writeIndexFile(o.Index(), in, nfiles)
		}
	}
}

// writeIndexFile writes Sitemap index file for N files.
func writeIndexFile(w io.Writer, in Input, nfiles int) error {
	start := xml.StartElement{
		Name: xml.Name{Local: "sitemapindex"},
		Attr: []xml.Attr{
			{
				Name:  xml.Name{Local: "xmlns"},
				Value: "http://www.sitemaps.org/schemas/sitemap/0.9",
			},
		},
	}

	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}

	e := xml.NewEncoder(w)
	e.Indent("", "  ")
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	sitemap := Url{}
	for i := 0; i < nfiles; i++ {
		sitemap.Loc = in.GetUrlsetUrl(i)
		if err := e.Encode(sitemap); err != nil {
			return err
		}
	}

	if err := e.EncodeToken(start.End()); err != nil {
		return err
	}

	return e.Flush()
}

// writeUrlsetFile writes a single Sitemap Urlset file for the first 50K entries
// in the given input.
func writeUrlsetFile(w io.Writer, in Input) (retErr error) {
	start := xml.StartElement{
		Name: xml.Name{Local: "urlset"},
		Attr: []xml.Attr{
			{
				Name:  xml.Name{Local: "xmlns"},
				Value: "http://www.sitemaps.org/schemas/sitemap/0.9",
			},
			{
				Name:  xml.Name{Local: "xmlns:image"},
				Value: "http://www.google.com/schemas/sitemap-image/1.1",
			},
		},
	}

	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}

	e := xml.NewEncoder(w)
	e.Indent("", "  ")
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	url := Url{}
	image := Image{}
	var count int
	for in.HasNext() {
		if count >= maxSitemapCap {
			retErr = errMaxCapReached{}
			break
		}

		entry := in.Next()

		url.Loc = entry.GetLoc()
		url.LastMod = t2str(entry.GetLastMod())
		url.Images = []Image{}
		for _, imageUrl := range entry.GetImages() {
			image.Loc = imageUrl
			url.Images = append(url.Images, image)
		}

		if err := e.Encode(url); err != nil {
			return err
		}

		count++
	}

	if err := e.EncodeToken(start.End()); err != nil {
		return err
	}

	if err := e.Flush(); err != nil {
		return err
	}

	return retErr
}

const (
	maxSitemapCap = 50_000
)

type errMaxCapReached struct{}

func (e errMaxCapReached) Error() string {
	return "max 50K capacity is reached"
}
