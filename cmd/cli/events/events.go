package main

import (
	"context"
	"os"
	"os/signal"
	"prevtorrent/internal/platform/container"
	"syscall"
)

func main() {
	c, err := container.NewDefaultContainer()
	if err != nil {
		panic(err)
	}

	router := c.CQRSRouter()

	ctx, cancelCtx := context.WithCancel(context.Background())
	go gracefulShutdown(cancelCtx)

	if err := router.Run(ctx); err != nil {
		panic(err)
	}
}

func gracefulShutdown(cancelCtx context.CancelFunc) {
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan
	cancelCtx()
}
