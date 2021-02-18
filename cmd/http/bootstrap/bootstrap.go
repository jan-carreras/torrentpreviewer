package bootstrap

import (
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/preview/platform/http"
)

func Run() error {
	c, err := http.NewContainer()
	if err != nil {
		return err
	}
	return http.Run(c, makeCommandBus(c))
}

func makeCommandBus(c http.Container) *inmemory.SyncCommandBus {
	return inmemory.NewSyncCommandBus(c.Logger)
}
