package main

import (
	"context"
	"prevtorrent/internal/platform/bus/pubsub"
	"prevtorrent/internal/preview/downloadPartials"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
)

func main() {
	logger := watermill.NewStdLogger(false, false)

	publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{ProjectID: "torrentpreview"}, logger)
	if err != nil {
		panic(err)
	}
	commandBus := pubsub.NewPubSubCommandBus(nil, publisher)

	cmd := downloadPartials.CMD{ID: "c92f656155d0d8e87d21471d7ea43e3ad0d42723"}

	if err := commandBus.Dispatch(context.Background(), cmd); err != nil {
		panic(err)
	}
}
