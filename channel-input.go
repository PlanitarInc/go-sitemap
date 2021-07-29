package sitemap

import (
	"strconv"
	"sync/atomic"
)

type ChannelInput struct {
	channel       chan UrlEntry
	closed        int32
	lastReadEntry UrlEntry
	baseUrl       string
	fileName       string
	extension     string
}

func NewChannelInput() *ChannelInput {
	return &ChannelInput{
		channel: make(chan UrlEntry),
	}
}

func (in *ChannelInput) Feed(entry UrlEntry) {
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

	atomic.StoreInt32(&in.closed, 1)
	close(in.channel)
}

func (in *ChannelInput) HasNext() bool {
	entry, ok := <-in.channel
	if !ok {
		in.lastReadEntry = nil
		return false
	}

	in.lastReadEntry = entry
	return true
}

func (in *ChannelInput) Next() UrlEntry {
	return in.lastReadEntry
}

func (in *ChannelInput) SetIndexUrl(baseUrl string, fileName string, extension string) {
	in.baseUrl = baseUrl
	in.fileName = fileName
	in.extension = extension
}

func (in *ChannelInput) GetIndexUrl(idx int) string {
	return in.baseUrl + in.fileName + strconv.Itoa(idx+1) + "." + in.extension
}
