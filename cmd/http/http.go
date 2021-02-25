package main

import (
	"log"
	"os"
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

	s, err := services.NewServices(c)
	if err != nil {
		return err
	}

	return http.Run(s)
}
