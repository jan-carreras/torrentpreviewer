package preview

type TorrentCreatedEvent struct {
	TorrentID string
}

func NewTorrentCreatedEvent(torrentID string) *TorrentCreatedEvent {
	return &TorrentCreatedEvent{TorrentID: torrentID}
}
