package main

import (
	"log"
	"os"
	"prevtorrent/internal/platform/container"
	"prevtorrent/internal/preview/platform/cli"
)

func main() {
	c, err := container.NewDefaultContainer()
	if err != nil {
		log.Fatal(err)
	}

	err = cli.Run(os.Args, c.CommandBus())
	if err != nil {
		log.Fatal(err)
	}
}
