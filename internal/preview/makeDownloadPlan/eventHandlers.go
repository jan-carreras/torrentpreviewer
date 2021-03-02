package makeDownloadPlan

import (
	"context"
	"prevtorrent/internal/preview"
)

type TorrentCreatedEventHandler struct {
	service Service
}

func NewTorrentCreatedEventHandler(service Service) *TorrentCreatedEventHandler {
	return &TorrentCreatedEventHandler{service: service}
}

func (b TorrentCreatedEventHandler) HandlerName() string {
	return "event.torrent.makeDownloadPlan"
}

func (TorrentCreatedEventHandler) NewEvent() interface{} {
	return new(preview.TorrentCreatedEvent)
}

func (b *TorrentCreatedEventHandler) Handle(ctx context.Context, e interface{}) error {
	event := e.(*preview.TorrentCreatedEvent)

	return b.service.Download(ctx, CMD{
		TorrentID: event.TorrentID,
	})
}
