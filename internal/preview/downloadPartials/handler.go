package downloadPartials

import (
	"context"
)

type CommandHandler struct {
	service Service
}

func NewCommandHandler(service Service) *CommandHandler {
	return &CommandHandler{service: service}
}

func (h CommandHandler) HandlerName() string {
	return "command.torrent.downloadPartials"
}

func (h CommandHandler) NewCommand() interface{} {
	return new(CMD)
}

func (h CommandHandler) Handle(ctx context.Context, c interface{}) error {
	return h.service.DownloadPartials(ctx, *c.(*CMD))
}
