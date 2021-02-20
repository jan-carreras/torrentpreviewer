package bootstrap

import (
	"encoding/json"
	"fmt"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/preview/platform/http"
)

func Run() error {
	c, err := http.NewContainer()
	if err != nil {
		return err
	}

	if conf, err := json.MarshalIndent(c.Config, "", "  "); err != nil {
		return err
	} else {
		fmt.Println("Configuration:")
		fmt.Println(string(conf))
	}

	return http.Run(c, makeCommandBus(c))
}

func makeCommandBus(c http.Container) *inmemory.SyncCommandBus {
	return inmemory.NewSyncCommandBus(c.Logger)
}
