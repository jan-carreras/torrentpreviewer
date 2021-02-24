package main

import (
	"context"
	"os"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/platform/container"
	"prevtorrent/internal/preview/downloadPartials"
)

func main() {
	args := os.Args
	torrentID := "c92f656155d0d8e87d21471d7ea43e3ad0d42723"
	if len(args) >= 2 {
		torrentID = args[1]
	}

	c, err := container.NewDefaultContainer()
	if err != nil {
		panic(err)
	}

	commandBus, err := bus.MakeCommandBus(bus.PubSub, c)
	if err != nil {
		panic(err)
	}

	cmd := downloadPartials.CMD{ID: torrentID}

	if err := commandBus.Dispatch(context.Background(), cmd); err != nil {
		panic(err)
	}
}
