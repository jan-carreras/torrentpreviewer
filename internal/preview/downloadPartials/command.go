package downloadPartials

import (
	"context"
	"errors"
	"prevtorrent/kit/command"
)

const CommandType command.Type = "command.downloadPartials.preview"

type CMD struct {
	ID string
}

func (c CMD) Type() command.Type {
	return CommandType
}

type CommandHandler struct {
	service Service
	cmd     CMD
}

func NewCommandHandler(service Service) CommandHandler {
	return CommandHandler{service: service}
}

func (c CommandHandler) Handle(ctx context.Context, _cmd command.Command) error {
	cmd, ok := _cmd.(CMD)
	if !ok {
		return errors.New("unexpected command")
	}

	return c.service.DownloadPartials(ctx, cmd)
}