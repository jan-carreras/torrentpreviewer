package main

import (
	"log"
	"os"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/platform/container"
	"prevtorrent/internal/preview/platform/cli"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	c, err := container.NewDefaultContainer()
	if err != nil {
		return err
	}

	commandBus, err := bus.MakeCommandBus(bus.Sync, c)
	if err != nil {
		return err
	}
	c.Config.Print(os.Stdout)
	return cli.Run(commandBus)
}
