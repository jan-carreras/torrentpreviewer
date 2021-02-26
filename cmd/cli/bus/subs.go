package main

import (
	"context"
	"prevtorrent/internal/platform/container"
)

func main() {
	c, err := container.NewDefaultContainer()
	if err != nil {
		panic(err)
	}

	_ = c.CQRS()

	router := c.CQRSRouter()
	if err := router.Run(context.Background()); err != nil {
		panic(err)
	}
}
