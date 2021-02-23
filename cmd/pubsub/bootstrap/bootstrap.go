package bootstrap

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/unmagnetize"
)

func Run(ctx context.Context) error {
	container, err := newContainer()
	if err != nil {
		return err
	}
	commandBus := makeCommandBus(container)
	sub := newSubscriber(ctx, commandBus, container.subscriber)

	err = sub.Listen(ctx)
	if err != nil {
		return err
	}

	return nil
}

type subscriber struct {
	ctx        context.Context
	commandBus *inmemory.SyncCommandBus
	subscriber *googlecloud.Subscriber
}

func newSubscriber(ctx context.Context, commandBus *inmemory.SyncCommandBus, sub *googlecloud.Subscriber) *subscriber {
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

func makeCommandBus(c container) *inmemory.SyncCommandBus {
	commandBus := inmemory.NewSyncCommandBus(c.logger)

	commandBus.Register(
		unmagnetize.CommandType,
		unmagnetize.NewTransformCommandHandler(
			unmagnetize.NewService(
				c.logger,
				c.magnetClient,
				c.torrentRepo,
			),
		),
	)

	commandBus.Register(
		downloadPartials.CommandType,
		downloadPartials.NewCommandHandler(
			downloadPartials.NewService(
				c.logger,
				c.torrentRepo,
				c.torrentDownloader,
				c.imageExtractor,
				c.imagePersister,
				c.imageRepository,
			),
		),
	)

	return commandBus
}
