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
			c.Logger.Error(ctx.Err())
			return ctx.Err()
		case msg, isOpen := <-downloadPartialsChannel:
			if !isOpen {
				c.Logger.Debug("downloadPartialsChannel has been closed")
				return nil
			}

			cmd := downloadPartials.CMD{}
			err := json.Unmarshal(msg.Payload(), &cmd)
			if err != nil {
				c.Logger.Error("unable to unmarshal payload", string(msg.Payload()))
				continue
			}
			err = commandBus.Dispatch(ctx, cmd)
			if err != nil {
				c.Logger.Error("error when processing Command", cmd)
				continue
			}
			msg.ACK()
		}
	}
}
