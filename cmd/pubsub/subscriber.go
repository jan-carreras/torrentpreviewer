package main

import (
	"context"
	"log"
	"prevtorrent/cmd/pubsub/bootstrap"
)

func main() {
	if err := bootstrap.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
