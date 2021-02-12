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
	// TODO: Do we really have to do that? It cannot be abstracted away?
	// TODO: Maybe "extending" and overwritting from an object that already has this method, maybe? Please!
	return CommandType
}

type CommandHandler struct {
	service Service
}

func NewCommandHandler(service Service) CommandHandler {
	return CommandHandler{service: service}
}

func (c CommandHandler) Handle(ctx context.Context, _cmd command.Command) error {
	// TODO: This is shite. Cannot be abstracted away, pretty please?
	// TODO:   do we **really** have to do it in every fucking ApplicationService?
	cmd, ok := _cmd.(CMD)
	if !ok {
		return errors.New("unexpected command")
	}

	return c.service.DownloadPartials(ctx, cmd)
}
