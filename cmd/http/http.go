package main

import (
	"log"
	"os"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/platform/container"
	"prevtorrent/internal/platform/services"
	"prevtorrent/internal/preview/platform/http"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	c, err := container.NewDefaultContainer()
	if err != nil {
		return err
	}

	c.Config.Print(os.Stdout)

	commandBus := inmemory.NewSyncCommandBus(c.Logger)
	bus.MakeBindings(commandBus, c)

	s, err := services.NewServices(c)
	if err != nil {
		return err
	}

	return http.Run(s, commandBus)
}
