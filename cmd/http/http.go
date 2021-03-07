package main

import (
	"context"
	"log"
	http2 "net/http"
	"os"
	"os/signal"
	"prevtorrent/internal/platform/container"
	"prevtorrent/internal/platform/services"
	"prevtorrent/internal/preview/platform/http"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
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

	c.Config().Print(os.Stdout)

	s, err := services.NewServices(c)
	if err != nil {
		return err
	}

	srv := http.Run(s)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http2.ErrServerClosed {
			c.Logger().Errorf("listen: %s\n", err)
		}
	}()

	waitForGracefulShutdown(c.Logger(), srv)

	return nil
}

func waitForGracefulShutdown(logger *logrus.Logger, srv *http2.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}
}
