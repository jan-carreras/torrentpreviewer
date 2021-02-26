package cli

import (
	"context"
	"errors"
	"github.com/urfave/cli/v2"
	"os"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/unmagnetize"
)

type handlers struct {
	commandBus bus.Command
}

func newHandlers(bus bus.Command) *handlers {
	return &handlers{commandBus: bus}
}

func Run(bus bus.Command) error {
	handlers := newHandlers(bus)

	app := &cli.App{
		Name:  "torrentprev",
		Usage: "your favourite torrent previewer",
		Commands: []*cli.Command{
			{
				Name:  "download",
				Usage: "download the given torrent ID - must have been imported first",
				Action: func(c *cli.Context) error {
					return handlers.download(c)
				},
			},
			{
				Name:  "magnet",
				Usage: "transforms a magnet link into a torrent and imports it",
				Action: func(c *cli.Context) error {
					return handlers.transform(c)
				},
			},
		},
	}

	return app.Run(os.Args)
}

func (h *handlers) transform(c *cli.Context) error {
	if c.NArg() != 1 {
		return errors.New("invalid arguments. Second parameter must be a magnet")
	}
	magnet := c.Args().Get(0)

	return h.commandBus.Send(context.Background(), unmagnetize.CMD{
		Magnet: magnet,
	})
}

func (h *handlers) download(c *cli.Context) error {
	if c.NArg() != 1 {
		return errors.New("second parameter must be the path to a valid torrent")
	}
	torrent := c.Args().Get(0)

	return h.commandBus.Send(context.Background(), downloadPartials.CMD{
		ID: torrent,
	})
}
