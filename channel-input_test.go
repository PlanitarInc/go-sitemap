package sitemap

import (
	"bytes"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestChannelFeed(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput()

		go in.Feed(nil)
		Eventually(in.channel).Should(Receive())
		Ω(in.channel).ShouldNot(BeClosed())
	})

	t.Run("entry", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput()
		go in.Feed(&SimpleEntry{Loc: "one"})
		Eventually(in.channel).Should(Receive(Equal(&SimpleEntry{Loc: "one"})))
		Ω(in.channel).ShouldNot(BeClosed())
	})
}

func TestChannelClose(t *testing.T) {
	RegisterTestingT(t)

	in := NewChannelInput()

	Ω(in.channel).ShouldNot(BeClosed())
	in.Close()
	Ω(in.channel).Should(BeClosed())
}

func TestChannelInputNext(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput()
		Ω(in.Next()).Should(BeNil())
	})

	t.Run("empty", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput()
		in.lastReadEntry = &SimpleEntry{Loc: "one"}

		Ω(in.Next()).Should(Equal(&SimpleEntry{Loc: "one"}))
	})
}

func TestChannelInputHasNext(t *testing.T) {
	t.Run("Feed", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput()

		Ω(in.lastReadEntry).Should(BeNil())
		go in.Feed(&SimpleEntry{Loc: "one"})
		Ω(in.HasNext()).Should(BeTrue())
		Ω(in.lastReadEntry).Should(Equal(&SimpleEntry{Loc: "one"}))
	})

	t.Run("Close", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput()
		in.lastReadEntry = &SimpleEntry{Loc: "one"}

		go func(in *ChannelInput) {
			time.Sleep(100 * time.Millisecond)
			in.Close()
		}(in)
		Ω(in.HasNext()).Should(BeFalse())
		Ω(in.lastReadEntry).Should(BeNil())
		Ω(in.channel).Should(BeClosed())
	})
}

func TestSitemapWriteChannelInput(t *testing.T) {
	RegisterTestingT(t)

	out := &bytes.Buffer{}
	in := NewChannelInput()

	go func(in *ChannelInput) {
		in.Feed(&SimpleEntry{Loc: "a"})
	}(in)
	go func(in *ChannelInput) {
		time.Sleep(100 * time.Millisecond)
		in.Feed(&SimpleEntry{Loc: "b"})
	}(in)
	go func(in *ChannelInput) {
		time.Sleep(200 * time.Millisecond)
		in.Feed(&SimpleEntry{Loc: "c"})
	}(in)
	go func(in *ChannelInput) {
		time.Sleep(500 * time.Millisecond)
		in.Close()
	}(in)

	Ω(SitemapWrite(out, in)).Should(BeNil())
	Ω(out.String()).Should(MatchXML(`
		<?xml version="1.0" encoding="UTF-8"?>
		<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"
			xmlns:image="http://www.google.com/schemas/sitemap-image/1.1">
			<url> <loc>a</loc> </url>
			<url> <loc>b</loc> </url>
			<url> <loc>c</loc> </url>
		</urlset>
	`))
}
