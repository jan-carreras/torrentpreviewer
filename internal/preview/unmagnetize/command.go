package unmagnetize

import (
	"context"
	"errors"
	"prevtorrent/kit/command"
	"time"
)

const CommandType command.Type = "command.transform.preview"

type CMD struct {
	Magnet string
}

func (c CMD) Type() command.Type {
	return CommandType
}

type CommandHandler struct {
	service Service
}

func NewTransformCommandHandler(service Service) CommandHandler {
	return CommandHandler{service: service}
}

func (c CommandHandler) Handle(ctx context.Context, _cmd command.Command) error {
	cmd, ok := _cmd.(CMD)
	if !ok {
		return errors.New("unexpected command")
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	return c.service.Handle(ctxTimeout, cmd)
}
