package bootstrap

import (
	"context"
	"encoding/json"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/platform/bus/inmemory"
	container2 "prevtorrent/internal/platform/container"
	"prevtorrent/internal/preview/downloadPartials"
)

func Run(ctx context.Context) error {
	container, err := container2.NewDefaultContainer()
	if err != nil {
		return err
	}

	commandBus := inmemory.NewSyncCommandBus(container.Logger)

	bus.MakeBindings(commandBus, container) // TODO: Meh.

	downloadPartialsChannel, err := container.Subscriber.Subscribe(ctx, string(downloadPartials.CommandType))
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, isOpen := <-downloadPartialsChannel:
			if !isOpen {
				return nil
			}

			cmd := downloadPartials.CMD{}
			err := json.Unmarshal(msg.Payload(), &cmd)
			if err != nil {
				return err
			}
			err = commandBus.Dispatch(ctx, cmd)
			if err != nil {
				return err
			}
			msg.ACK()
		}
	}
}
