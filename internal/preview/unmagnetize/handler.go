package unmagnetize

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
	return "command.magnet.unmagnetize"
}

func (h CommandHandler) NewCommand() interface{} {
	return new(CMD)
}

func (h CommandHandler) Handle(ctx context.Context, c interface{}) error {
	_, err := h.service.Handle(ctx, *c.(*CMD))
	return err
}
