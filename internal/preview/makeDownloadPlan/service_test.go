package makeDownloadPlan_test

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"prevtorrent/internal/platform/bus/busmocks"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/makeDownloadPlan"
	"prevtorrent/internal/preview/platform/storage/storagemocks"
	"testing"
)

func TestService_Download_ErrorOnGettingTorrent(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, torrentID).
		Return(preview.Torrent{}, errors.New("fake error"))

	imageRepository := new(storagemocks.ImageRepository)

	commandBus := new(busmocks.Command)

	service := makeDownloadPlan.NewService(fakeLogger(), commandBus, torrentRepository, imageRepository)

	err := service.Download(context.Background(), makeDownloadPlan.CMD{
		TorrentID: torrentID,
	})

	require.Error(t, err)
}

func TestService_Download_ErrorOnGettingTorrentImages(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"

	f, err := preview.NewFileInfo(0, 10, "video.mp4")
	require.NoError(t, err)

	files := []preview.File{f}

	torrent, err := preview.NewInfo(torrentID, "test torrent", 5, files, []byte("torrent-data"))
	require.NoError(t, err)

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, torrentID).
		Return(torrent, nil)

	imageRepository := new(storagemocks.ImageRepository)
	imageRepository.On("ByTorrent", mock.Anything, torrentID).
		Return(nil, errors.New("fake read images error"))

	commandBus := new(busmocks.Command)

	service := makeDownloadPlan.NewService(fakeLogger(), commandBus, torrentRepository, imageRepository)

	err = service.Download(context.Background(), makeDownloadPlan.CMD{
		TorrentID: torrentID,
	})

	require.Error(t, err)
}

func TestService_Download_ErrorOnExecutingCommands(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"

	f, err := preview.NewFileInfo(0, 10, "video.mp4")
	require.NoError(t, err)

	files := []preview.File{f}

	torrent, err := preview.NewInfo(torrentID, "test torrent", 5, files, []byte("torrent-data"))
	require.NoError(t, err)

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, torrentID).
		Return(torrent, nil)

	imageRepository := new(storagemocks.ImageRepository)
	imageRepository.On("ByTorrent", mock.Anything, torrentID).
		Return(new(preview.TorrentImages), nil)

	commandBus := new(busmocks.Command)
	commandBus.On("Send", mock.Anything, mock.Anything).
		Return(errors.New("fake publish error"))

	service := makeDownloadPlan.NewService(fakeLogger(), commandBus, torrentRepository, imageRepository)

	err = service.Download(context.Background(), makeDownloadPlan.CMD{
		TorrentID: torrentID,
	})

	require.Error(t, err)
}

func TestService_Download_BaseCase(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"

	f, err := preview.NewFileInfo(0, 10, "video.mp4")
	require.NoError(t, err)

	files := []preview.File{f}

	torrent, err := preview.NewInfo(torrentID, "test torrent", 5, files, []byte("torrent-data"))
	require.NoError(t, err)

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, torrentID).
		Return(torrent, nil)

	imageRepository := new(storagemocks.ImageRepository)
	imageRepository.On("ByTorrent", mock.Anything, torrentID).
		Return(new(preview.TorrentImages), nil)

	commandBus := new(busmocks.Command)
	commandBus.On("Send", mock.Anything, mock.Anything).
		Return(nil)

	service := makeDownloadPlan.NewService(fakeLogger(), commandBus, torrentRepository, imageRepository)

	err = service.Download(context.Background(), makeDownloadPlan.CMD{
		TorrentID: torrentID,
	})

	require.NoError(t, err)
}

func fakeLogger() *logrus.Logger {
	l := logrus.New()
	l.Out = ioutil.Discard
	return l
}
