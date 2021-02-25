package pubsub

import (
	"context"
	"encoding/json"
	"prevtorrent/kit/command"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/sirupsen/logrus"
)

type PubSubCommandBus struct {
	logger    *logrus.Logger
	publisher message.Publisher
}

func NewPubSubCommandBus(logger *logrus.Logger, publisher message.Publisher) *PubSubCommandBus {
	return &PubSubCommandBus{
		logger:    logger,
		publisher: publisher,
	}
}

func (b *PubSubCommandBus) Register(_ command.Type, _ command.Handler) {
}

func (b *PubSubCommandBus) Dispatch(_ context.Context, cmd command.Command) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	msg := message.NewMessage(watermill.NewUUID(), data)
	return b.publisher.Publish(string(cmd.Type()), msg)
}
