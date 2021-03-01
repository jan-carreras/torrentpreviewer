package downloadPartials

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
	return "event.torrent.downloadPartials"
}

func (TorrentCreatedEventHandler) NewEvent() interface{} {
	return &preview.TorrentCreatedEvent{}
}

func (b *TorrentCreatedEventHandler) Handle(ctx context.Context, e interface{}) error {
	event := e.(*preview.TorrentCreatedEvent)

	// TODO: Migrate this event handler to makeDownloadPlan!!!

	return b.service.DownloadPartials(ctx, CMD{
		ID: event.TorrentID,
	})
}
