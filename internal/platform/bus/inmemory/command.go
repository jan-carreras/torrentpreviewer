package inmemory

import (
	"context"
	"github.com/sirupsen/logrus"
	"prevtorrent/kit/command"
)

type SyncCommandBus struct {
	handlers map[command.Type]command.Handler
	logger   *logrus.Logger
}

func NewSyncCommandBus(logger *logrus.Logger) *SyncCommandBus {
	return &SyncCommandBus{
		handlers: make(map[command.Type]command.Handler),
		logger:   logger,
	}
}

func (b *SyncCommandBus) Dispatch(ctx context.Context, cmd command.Command) error {
	handler, ok := b.handlers[cmd.Type()]
	if !ok {
		b.logger.WithFields(logrus.Fields{
			"cmdType": cmd.Type(),
		}).Error("there is no handler registered for this command")
		return nil
	}

	err := handler.Handle(ctx, cmd)
	if err != nil {
		b.logger.WithFields(logrus.Fields{
			"cmdType": cmd.Type(),
			"err":     err.Error(),
		}).Error("error while handling a command")
	}

	return nil
}

func (b *SyncCommandBus) Register(cmdType command.Type, handler command.Handler) {
	b.handlers[cmdType] = handler
}
