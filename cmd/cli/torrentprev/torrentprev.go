package main

import (
	"log"
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

	return cli.Run(c.CQRS().CommandBus())
}
