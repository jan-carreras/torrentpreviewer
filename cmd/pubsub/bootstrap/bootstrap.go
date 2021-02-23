package bootstrap

import (
	"context"
	"encoding/json"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/platform/bus/inmemory"
	container2 "prevtorrent/internal/platform/container"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/kit/command"

	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
)

func Run(ctx context.Context) error {
	container, err := container2.NewDefaultContainer()
	if err != nil {
		return err
	}

	commandBus := inmemory.NewSyncCommandBus(container.Logger)

	bus.MakeBindings(commandBus, container)

	sub := newSubscriber(ctx, commandBus, container.Subscriber)

	err = sub.Listen(ctx)
	if err != nil {
		return err
	}

	return nil
}

type subscriber struct {
	ctx        context.Context
	commandBus command.Bus
	subscriber *googlecloud.Subscriber
}

func newSubscriber(ctx context.Context, commandBus command.Bus, sub *googlecloud.Subscriber) *subscriber {
	return &subscriber{ctx: ctx, commandBus: commandBus, subscriber: sub}
}

func (s subscriber) Listen(ctx context.Context) error {
	downloadPartialsChannel, err := s.subscriber.Subscribe(ctx, "command.downloadPartials")
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
			err := json.Unmarshal(msg.Payload, &cmd)
			if err != nil {
				return err
			}
			err = s.commandBus.Dispatch(ctx, cmd)
			if err != nil {
				return err
			}
			// we need to Acknowledge that we received and processed the message,
			// otherwise, it will be resent over and over again.
			msg.Ack()
		}
	}
}
