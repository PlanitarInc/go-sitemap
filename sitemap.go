package sitemap

import (
	"encoding/xml"
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

// XXX The limitation of file size:
//  - according to http://www.sitemaps.org/protocol.html:
//     <50.000 urls and <10MB
// XXX It should be fine until we reach 10.000 entries
func SitemapWrite(w io.Writer, in Input) error {
	urlset := xml.Name{Local: "urlset"}
	start := xml.StartElement{
		Name: urlset,
		Attr: []xml.Attr{
			xml.Attr{
				Name:  xml.Name{Local: "xmlns"},
				Value: "http://www.sitemaps.org/schemas/sitemap/0.9",
			},
			xml.Attr{
				Name:  xml.Name{Local: "xmlns:image"},
				Value: "http://www.google.com/schemas/sitemap-image/1.1",
			},
		},
	}

	io.WriteString(w, xml.Header)
	e := xml.NewEncoder(w)
	e.Indent("", "  ")

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	url := Url{}
	image := Image{}
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
	}

	if err := e.EncodeToken(start.End()); err != nil {
		return err
	}

	return e.Flush()
}
