package inmemory

import (
	"context"
	"fmt"
	"log"
	"prevtorrent/kit/command"
)

type SyncCommandBus struct {
	handlers map[command.Type]command.Handler
}

func NewSyncCommandBus() *SyncCommandBus {
	return &SyncCommandBus{
		handlers: make(map[command.Type]command.Handler),
	}
}

func (b *SyncCommandBus) Dispatch(ctx context.Context, cmd command.Command) error {
	handler, ok := b.handlers[cmd.Type()]
	if !ok {
		fmt.Println("oh, right... we don't find anything and we don't even fail :/")
		return nil
	}

	err := handler.Handle(ctx, cmd)
	if err != nil {
		log.Printf("Error while handling %s - %s\n", cmd.Type(), err)
	}

	return nil
}

func (b *SyncCommandBus) Register(cmdType command.Type, handler command.Handler) {
	b.handlers[cmdType] = handler
}
