package makeDownloadPlan

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
	return "command.torrent.makeDownloadPlan"
}

func (h CommandHandler) NewCommand() interface{} {
	return new(CMD)
}

func (h CommandHandler) Handle(ctx context.Context, c interface{}) error {
	return h.service.Download(ctx, *c.(*CMD))
}
