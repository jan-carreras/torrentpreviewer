package downloadPartials

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type CommandHandler struct {
	eventBus *cqrs.EventBus
	service  Service
}

func NewCommandHandler(eventBus *cqrs.EventBus, service Service) *CommandHandler {
	return &CommandHandler{eventBus: eventBus, service: service}
}

func (h CommandHandler) HandlerName() string {
	return "command.magnet.unmagnetize"
}

func (h CommandHandler) NewCommand() interface{} {
	return new(CMD)
}

func (h CommandHandler) Handle(ctx context.Context, c interface{}) error {
	return h.service.DownloadPartials(ctx, *c.(*CMD))
}
