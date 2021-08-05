package sitemap

import "sync/atomic"

type ChannelInput struct {
	channel      chan *UrlEntry
	closed       int32
	getUrlsetUrl func(int) string
}

func NewChannelInput(getUrlsetUrl func(int) string) *ChannelInput {
	return &ChannelInput{
		channel:      make(chan *UrlEntry),
		getUrlsetUrl: getUrlsetUrl,
	}
}

func (in *ChannelInput) Feed(entry *UrlEntry) {
	if atomic.LoadInt32(&in.closed) > 0 {
		return
	}

	in.channel <- entry
}

func (in *ChannelInput) Close() {
	if atomic.LoadInt32(&in.closed) > 0 {
		return
	}

	defer func() {
		_ = recover()
	}()

	if atomic.SwapInt32(&in.closed, 1) == 0 {
		close(in.channel)
	}
}

func (in *ChannelInput) Next() *UrlEntry {
	entry, ok := <-in.channel
	if !ok {
		return nil
	}

	return entry
}

func (in *ChannelInput) GetUrlsetUrl(n int) string {
	if in.getUrlsetUrl == nil {
		return ""
	}

	return in.getUrlsetUrl(n)
}
