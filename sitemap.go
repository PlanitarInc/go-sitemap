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

type Output interface {
	Index() io.Writer
	Urlset() io.Writer
}

func WriteWithIndex(o Output, in Input, max int) error {
	var nfiles int
	for {
		nfiles++
		err := writeUrlset(o.Urlset(), in, max)
		if err != nil && !errors.Is(err, errMaxCapReached{}) {
			return err
		}

		if err == nil {
			if nfiles > 1 {
				return writeIndex(o.Index(), in, nfiles)
			}

			break
		}
	}

	return nil
}

func writeIndex(w io.Writer, in Input, nfiles int) error {
	sitemapindex := xml.Name{Local: "sitemapindex"}
	start := xml.StartElement{
		Name: sitemapindex,
		Attr: []xml.Attr{
			xml.Attr{
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
		sitemap.Loc = in.GetIndexUrl(i)

		if err := e.Encode(sitemap); err != nil {
			return err
		}
	}

	if err := e.EncodeToken(start.End()); err != nil {
		return err
	}

	return e.Flush()
}

func writeUrlset(w io.Writer, in Input, max int) error {
	urlset := xml.Name{Local: "urlset"}
	start := xml.StartElement{
		Name: urlset,
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

	_, _ = io.WriteString(w, xml.Header)
	e := xml.NewEncoder(w)
	e.Indent("", "  ")

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	url := Url{}
	image := Image{}
	var count int
	for in.HasNext() {
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
		if count >= max {
			break
		}
	}

	if err := e.EncodeToken(start.End()); err != nil {
		return err
	}

	if err := e.Flush(); err != nil {
		return err
	}

	if count >= max {
		return errMaxCapReached{}
	}

	return nil
}

type errMaxCapReached struct{}

func (e errMaxCapReached) Error() string {
	return "Max 50K capacity is reached"
}
