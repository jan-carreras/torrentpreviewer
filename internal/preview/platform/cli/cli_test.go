package cli_test

import (
	"prevtorrent/internal/platform/bus/busmocks"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/platform/cli"
	"prevtorrent/internal/preview/unmagnetize"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTorrentPrev_MagnetEventIsTriggered(t *testing.T) {
	commandBus := new(busmocks.Command)
	commandBus.On("Send", mock.Anything, &unmagnetize.CMD{
		Magnet: "magnet:?xt=urn:btih:c92f656155d0d8e87d21471d7ea43e3ad0d42723",
	}).Return(nil)

	args := []string{
		"test",
		"magnet",
		"magnet:?xt=urn:btih:c92f656155d0d8e87d21471d7ea43e3ad0d42723",
	}

	err := cli.Run(args, commandBus)
	require.NoError(t, err)
}

func TestTorrentPrev_MagnetFailsOnInvalidArguments(t *testing.T) {
	commandBus := new(busmocks.Command)
	commandBus.On("Send", mock.Anything, &unmagnetize.CMD{
		Magnet: "magnet:?xt=urn:btih:c92f656155d0d8e87d21471d7ea43e3ad0d42723",
	}).Return(nil)

	args := []string{
		"test",
		"magnet",
	}

	err := cli.Run(args, commandBus)
	require.Error(t, err)
}

func TestTorrentPrev_DownloadEventIsTriggered(t *testing.T) {
	commandBus := new(busmocks.Command)
	commandBus.On("Send", mock.Anything, &downloadPartials.CMD{
		ID: "c92f656155d0d8e87d21471d7ea43e3ad0d42723",
	}).Return(nil)

	args := []string{
		"test",
		"download",
		"c92f656155d0d8e87d21471d7ea43e3ad0d42723",
	}

	err := cli.Run(args, commandBus)
	require.NoError(t, err)
}

func TestTorrentPrev_DownloadFailsOnInvalidArguments(t *testing.T) {
	commandBus := new(busmocks.Command)
	commandBus.On("Send", mock.Anything, &downloadPartials.CMD{
		ID: "c92f656155d0d8e87d21471d7ea43e3ad0d42723",
	}).Return(nil)

	args := []string{
		"test",
		"download",
	}

	err := cli.Run(args, commandBus)
	require.Error(t, err)
}
