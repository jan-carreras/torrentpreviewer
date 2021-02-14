package transform_test

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/client/clientmocks"
	"prevtorrent/internal/preview/platform/storage/storagemocks"
	"prevtorrent/internal/preview/transform"
	"testing"
)

func Test_MagnetService_Transform_DownloadByNetwork(t *testing.T) {
	inputMagnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"
	torrent, err := preview.NewInfo(
		"zocmzqipffw7ollmic5hub6bpcsdeoqu",
		"test torrent",
		100,
		10,
		nil,
		[]byte("torrent-data"),
	)
	assert.NoError(t, err)

	mag, err := preview.NewMagnet(inputMagnet)
	assert.NoError(t, err)

	resolverRepo := new(clientmocks.MagnetClient)
	resolverRepo.On("Resolve", mock.Anything, mag).Return(torrent, nil)

	torrentRepo := new(storagemocks.TorrentRepository)
	torrentRepo.On("Persist", mock.Anything, torrent).Return(nil)
	torrentRepo.On("Get", mock.Anything, "zocmzqipffw7ollmic5hub6bpcsdeoqu").Return(preview.Info{}, preview.ErrNotFound)

	s := transform.NewService(fakeLogger(), resolverRepo, torrentRepo)
	err = s.Handle(context.Background(), transform.CMD{Magnet: inputMagnet})
	assert.NoError(t, err)
}

func Test_MagnetService_Transform_AlreadyDownloaded(t *testing.T) {
	inputMagnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"
	torrentData := []byte("torrent-data")

	resolverRepo := new(clientmocks.MagnetClient)

	torrentRepo := new(storagemocks.TorrentRepository)
	torrentRepo.On("Persist", mock.Anything, torrentData).Return(nil)
	torrentRepo.On("Get", mock.Anything, "zocmzqipffw7ollmic5hub6bpcsdeoqu").Return(preview.Info{}, nil)

	s := transform.NewService(fakeLogger(), resolverRepo, torrentRepo)
	err := s.Handle(context.Background(), transform.CMD{Magnet: inputMagnet})
	assert.NoError(t, err)
}

func Test_MagnetService_Transform_RepositoryErrorGetTorrent(t *testing.T) {
	inputMagnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"
	torrentData := []byte("torrent-data")

	resolverRepo := new(clientmocks.MagnetClient)

	torrentRepo := new(storagemocks.TorrentRepository)
	torrentRepo.On("Persist", mock.Anything, torrentData).Return(nil)
	torrentRepo.On("Get", mock.Anything, "zocmzqipffw7ollmic5hub6bpcsdeoqu").Return(preview.Info{}, errors.New("fake error"))

	s := transform.NewService(fakeLogger(), resolverRepo, torrentRepo)
	err := s.Handle(context.Background(), transform.CMD{Magnet: inputMagnet})
	assert.Error(t, err)
}

func Test_MagnetService_Inspect_RepositoryError(t *testing.T) {
	inputMagnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"

	mag, err := preview.NewMagnet(inputMagnet)
	assert.NoError(t, err)

	torrent, err := preview.NewInfo(
		"zocmzqipffw7ollmic5hub6bpcsdeoqu",
		"test torrent",
		100,
		10,
		nil,
		[]byte("torrent-data"),
	)
	assert.NoError(t, err)

	resolverRepo := new(clientmocks.MagnetClient)
	resolverRepo.On("Resolve", mock.Anything, mag).
		Return(torrent, errors.New("fake repo error"))

	torrentRepo := new(storagemocks.TorrentRepository)
	torrentRepo.On("Get", mock.Anything, "zocmzqipffw7ollmic5hub6bpcsdeoqu").
		Return(preview.Info{}, preview.ErrNotFound)

	s := transform.NewService(fakeLogger(), resolverRepo, torrentRepo)
	err = s.Handle(context.Background(), transform.CMD{Magnet: inputMagnet})
	assert.Error(t, err)
}

func Test_MagnetService_Inspect_InvalidMagnetError(t *testing.T) {
	inputMagnet := "invalid magnet"

	resolverRepo := new(clientmocks.MagnetClient)
	torrentRepo := new(storagemocks.TorrentRepository)

	s := transform.NewService(fakeLogger(), resolverRepo, torrentRepo)
	err := s.Handle(context.Background(), transform.CMD{Magnet: inputMagnet})
	assert.Error(t, err)
}

func fakeLogger() *logrus.Logger {
	l := logrus.New()
	l.Out = ioutil.Discard
	return l
}
