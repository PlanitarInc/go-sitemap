package sitemap

type ChannelInput struct {
	channel       chan UrlEntry
	lastReadEntry UrlEntry
}

func NewChannelInput() *ChannelInput {
	return &ChannelInput{
		channel: make(chan UrlEntry),
	}
}

func (in *ChannelInput) Feed(entry UrlEntry) {
	in.channel <- entry
}

func (in *ChannelInput) Close() {
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
