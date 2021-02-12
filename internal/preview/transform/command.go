package transform

import (
	"context"
	"errors"
	"prevtorrent/kit/command"
)

const TransformCommandType command.Type = "command.transform.preview"

type ServiceCMD struct {
	Magnet string
}

func (c ServiceCMD) Type() command.Type {
	// TODO: Do we really have to do that? It cannot be abstracted away?
	// TODO: Maybe "extending" and overwritting from an object that already has this method, maybe? Please!
	return TransformCommandType
}

type TransformCommandHandler struct {
	service Service
}

func NewTransformCommandHandler(service Service) TransformCommandHandler {
	return TransformCommandHandler{service: service}
}

func (c TransformCommandHandler) Handle(ctx context.Context, _cmd command.Command) error {
	// TODO: This is shite. Cannot be abstracted away, pretty please?
	// TODO:   do we **really** have to do it in every fucking ApplicationService?
	cmd, ok := _cmd.(ServiceCMD)
	if !ok {
		return errors.New("unexpected command")
	}

	return c.service.Handle(ctx, cmd)
}
