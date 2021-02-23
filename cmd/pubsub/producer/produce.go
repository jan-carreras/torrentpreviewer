package main

import (
	"context"
	"prevtorrent/internal/platform/bus/pubsub"
	"prevtorrent/internal/preview/downloadPartials"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/ThreeDotsLabs/watermill/message"
)

func main() {
	logger := watermill.NewStdLogger(false, false)

	publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{
		ProjectID: "torrentpreview",
	}, logger)
	if err != nil {
		panic(err)
	}

	publishMessages(publisher)
}

func publishMessages(publisher message.Publisher) {
	commandBus := pubsub.NewPubSubCommandBus(nil, publisher)
	cmd := downloadPartials.CMD{
		ID: "c92f656155d0d8e87d21471d7ea43e3ad0d42723",
	}
	err := commandBus.Dispatch(context.Background(), cmd)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second)
}
