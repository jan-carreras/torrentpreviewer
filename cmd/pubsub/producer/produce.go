package main

import (
	"context"
	"log"
	"os"
	"prevtorrent/internal/platform/bus/pubsub"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/platform/configuration"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
)

func main() {
	args := os.Args
	torrentID := "c92f656155d0d8e87d21471d7ea43e3ad0d42723"
	if len(args) >= 2 {
		torrentID = args[1]
	}

	logger := watermill.NewStdLogger(false, false)

	config, err := configuration.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{ProjectID: config.GooglePubSubProjectID}, logger)
	if err != nil {
		panic(err)
	}
	commandBus := pubsub.NewPubSubCommandBus(nil, publisher)

	cmd := downloadPartials.CMD{ID: torrentID}

	if err := commandBus.Dispatch(context.Background(), cmd); err != nil {
		panic(err)
	}
}
