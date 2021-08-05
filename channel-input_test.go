package sitemap

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestChannelInput_Close(t *testing.T) {
	t.Run("close", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)

		Ω(in.channel).ShouldNot(BeClosed())
		Ω(in.closed).Should(BeEquivalentTo(0))
		in.Close()
		Ω(in.channel).Should(BeClosed())
		Ω(in.closed).Should(BeNumerically(">", 0))
	})

	t.Run("doubleClose", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)

		Ω(in.channel).ShouldNot(BeClosed())
		Ω(in.closed).Should(BeEquivalentTo(0))
		in.Close()
		Ω(in.channel).Should(BeClosed())
		Ω(in.closed).Should(BeNumerically(">", 0))
		in.Close()
		Ω(in.closed).Should(BeNumerically(">", 0))
	})
}

func TestChannelInput_Feed(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)

		go in.Feed(nil)
		Eventually(in.channel).Should(Receive())
		Ω(in.channel).ShouldNot(BeClosed())
	})

	t.Run("entry", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)
		go in.Feed(&UrlEntry{Loc: "one"})
		Eventually(in.channel).Should(Receive(Equal(&UrlEntry{Loc: "one"})))
		Ω(in.channel).ShouldNot(BeClosed())
	})

	t.Run("closedChannel", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)
		in.Close()
		Ω(in.channel).Should(BeClosed())
		in.Feed(&UrlEntry{Loc: "one"})
	})
}

func TestChannelInput_Next(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)
		Ω(in.Next()).Should(BeNil())
	})

	t.Run("next", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)
		in.lastReadEntry = &UrlEntry{Loc: "one"}

		Ω(in.Next()).Should(Equal(&UrlEntry{Loc: "one"}))
	})
}

func TestChannelInput_HasNext(t *testing.T) {
	t.Run("Feed", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)

		Ω(in.lastReadEntry).Should(BeNil())
		go in.Feed(&UrlEntry{Loc: "one"})
		Ω(in.HasNext()).Should(BeTrue())
		Ω(in.lastReadEntry).Should(Equal(&UrlEntry{Loc: "one"}))
	})

	t.Run("Close", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)
		in.lastReadEntry = &UrlEntry{Loc: "one"}

		go func(in *ChannelInput) {
			time.Sleep(100 * time.Millisecond)
			in.Close()
		}(in)
		Ω(in.HasNext()).Should(BeFalse())
		Ω(in.lastReadEntry).Should(BeNil())
		Ω(in.channel).Should(BeClosed())
	})
}

func TestChannelInput_GetUrlsetUrl(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)
		Ω(in.GetUrlsetUrl(21)).Should(Equal(""))
	})

	t.Run("customUrl", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(func(n int) string {
			return fmt.Sprintf("@%d@", n)
		})

		Ω(in.GetUrlsetUrl(31)).Should(Equal("@31@"))
	})
}

func TestWriteAll_ChannelInput(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		RegisterTestingT(t)

		var out bufferOuput
		in := NewChannelInput(func(idx int) string {
			return fmt.Sprintf("channel input urlset %d", idx)
		})

		go func(in *ChannelInput) {
			in.Feed(&UrlEntry{Loc: "a"})
			in.Feed(&UrlEntry{Loc: "b"})
			in.Feed(&UrlEntry{Loc: "c"})
			in.Close()
		}(in)

		Ω(WriteAll(&out, in)).Should(BeNil())

		Ω(out.index.String()).Should(MatchXML(`
			<?xml version="1.0" encoding="UTF-8"?>
			<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
			  <url>
				<loc>channel input urlset 0</loc>
			  </url>
			</sitemapindex>
		`))

		Ω(out.sitemaps).Should(HaveLen(1))
		Ω(out.sitemaps[0].String()).Should(MatchXML(`
			<?xml version="1.0" encoding="UTF-8"?>
			<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
				xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
				<url> <loc>a</loc> </url>
				<url> <loc>b</loc> </url>
				<url> <loc>c</loc> </url>
			</urlset>
		`))
	})

	t.Run("concurrent", func(t *testing.T) {
		RegisterTestingT(t)

		var out bufferOuput
		in := NewChannelInput(func(idx int) string {
			return fmt.Sprintf("channel input urlset %d", idx)
		})

		go func(in *ChannelInput) {
			in.Feed(&UrlEntry{Loc: "a"})
		}(in)
		go func(in *ChannelInput) {
			time.Sleep(100 * time.Millisecond)
			in.Feed(&UrlEntry{Loc: "b"})
		}(in)
		go func(in *ChannelInput) {
			time.Sleep(200 * time.Millisecond)
			in.Feed(&UrlEntry{Loc: "c"})
		}(in)
		go func(in *ChannelInput) {
			time.Sleep(500 * time.Millisecond)
			in.Close()
		}(in)

		Ω(WriteAll(&out, in)).Should(BeNil())

		Ω(out.index.String()).Should(MatchXML(`
			<?xml version="1.0" encoding="UTF-8"?>
			<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
			  <url>
				<loc>channel input urlset 0</loc>
			  </url>
			</sitemapindex>
		`))

		Ω(out.sitemaps).Should(HaveLen(1))
		Ω(out.sitemaps[0].String()).Should(MatchXML(`
			<?xml version="1.0" encoding="UTF-8"?>
			<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
				xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
				<url> <loc>a</loc> </url>
				<url> <loc>b</loc> </url>
				<url> <loc>c</loc> </url>
			</urlset>
		`))
	})

	t.Run("multiplePages", func(t *testing.T) {
		RegisterTestingT(t)

		inputSize := 58_765
		var out bufferOuput
		in := NewChannelInput(func(idx int) string {
			return fmt.Sprintf("urlset %03d", idx)
		})

		go func(in *ChannelInput) {
			for i := 0; i < inputSize; i++ {
				in.Feed(&UrlEntry{Loc: fmt.Sprintf("http://goiguide.com/%d", i+1)})
			}
			in.Close()
		}(in)

		Ω(WriteAll(&out, in)).Should(BeNil())
		assertOutput(&out, inputSize)
	})
}
