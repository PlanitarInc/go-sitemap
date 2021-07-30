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
		go in.Feed(&simpleEntry{Loc: "one"})
		Eventually(in.channel).Should(Receive(Equal(&simpleEntry{Loc: "one"})))
		Ω(in.channel).ShouldNot(BeClosed())
	})

	t.Run("closedChannel", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)
		in.Close()
		Ω(in.channel).Should(BeClosed())
		in.Feed(&simpleEntry{Loc: "one"})
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
		in.lastReadEntry = &simpleEntry{Loc: "one"}

		Ω(in.Next()).Should(Equal(&simpleEntry{Loc: "one"}))
	})
}

func TestChannelInput_HasNext(t *testing.T) {
	t.Run("Feed", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)

		Ω(in.lastReadEntry).Should(BeNil())
		go in.Feed(&simpleEntry{Loc: "one"})
		Ω(in.HasNext()).Should(BeTrue())
		Ω(in.lastReadEntry).Should(Equal(&simpleEntry{Loc: "one"}))
	})

	t.Run("Close", func(t *testing.T) {
		RegisterTestingT(t)

		in := NewChannelInput(nil)
		in.lastReadEntry = &simpleEntry{Loc: "one"}

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
	RegisterTestingT(t)

	var out bufferOuput
	in := NewChannelInput(func(idx int) string {
		return fmt.Sprintf("channel input urlset %d", idx)
	})

	go func(in *ChannelInput) {
		in.Feed(&simpleEntry{Loc: "a"})
	}(in)
	go func(in *ChannelInput) {
		time.Sleep(100 * time.Millisecond)
		in.Feed(&simpleEntry{Loc: "b"})
	}(in)
	go func(in *ChannelInput) {
		time.Sleep(200 * time.Millisecond)
		in.Feed(&simpleEntry{Loc: "c"})
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
}
