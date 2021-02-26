package downloadPartials

import (
	"context"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/preview"
)

type TorrentCreatedEventHandler struct {
	eventBus bus.Event
	service  Service
}

func NewTorrentCreatedEventHandler(eventBus bus.Event, service Service) *TorrentCreatedEventHandler {
	return &TorrentCreatedEventHandler{eventBus: eventBus, service: service}
}

func (b TorrentCreatedEventHandler) HandlerName() string {
	return "event.torrent.downloadPartials"
}

func (TorrentCreatedEventHandler) NewEvent() interface{} {
	return &preview.TorrentCreatedEvent{}
}

func (b *TorrentCreatedEventHandler) Handle(ctx context.Context, e interface{}) error {
	event := e.(*preview.TorrentCreatedEvent)

	return b.service.DownloadPartials(ctx, CMD{
		ID: event.TorrentID,
	})
}
