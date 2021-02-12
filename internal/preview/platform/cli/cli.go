package cli

import (
	"context"
	"errors"
	"os"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/transform"
	"prevtorrent/kit/command"
)

/**
IMPROVEMENT: Use spf13/cobra or urfave/cli
*/

func Run(bus command.Bus) error {
	action := "unknown"
	if len(os.Args) >= 2 {
		action = os.Args[1]
	}

	if action == "download" {
		return download(bus)
	} else if action == "transform" {
		return trans(bus)
	} else {
		return errors.New("unknown command")
	}
}

func trans(bus command.Bus) error {
	if len(os.Args) != 3 {
		return errors.New("invalid arguments. Second parameter must be a magnet")
	}
	magnet := os.Args[2]

	return bus.Dispatch(context.Background(), transform.CMD{
		Magnet: magnet,
	})
}

func download(bus command.Bus) error {
	if len(os.Args) != 3 {
		return errors.New("second parameter must be the path to a valid torrent")
	}
	torr := os.Args[2]
	return bus.Dispatch(context.Background(), downloadPartials.CMD{ID: torr})
}
