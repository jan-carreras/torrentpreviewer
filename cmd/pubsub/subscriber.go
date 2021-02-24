package main

import (
	"context"
	"encoding/json"
	"log"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/platform/container"
	"prevtorrent/internal/preview/downloadPartials"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	c, err := container.NewDefaultContainer()
	if err != nil {
		return err
	}

	commandBus := inmemory.NewSyncCommandBus(c.Logger)

	bus.MakeBindings(commandBus, c) // TODO: Meh.

	downloadPartialsChannel, err := c.Subscriber.Subscribe(ctx, string(downloadPartials.CommandType))
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
