package unmagnetize_test

import (
	"context"
	"errors"
	"io/ioutil"
	"prevtorrent/internal/platform/bus/busmocks"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/client/clientmocks"
	"prevtorrent/internal/preview/platform/storage/storagemocks"
	"prevtorrent/internal/preview/unmagnetize"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_MagnetService_Transform_DownloadByNetwork(t *testing.T) {
	inputMagnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"
	torrent, err := preview.NewInfo(
		"zocmzqipffw7ollmic5hub6bpcsdeoqu",
		"test torrent",
		100,
		nil,
		[]byte("torrent-data"),
	)
	require.NoError(t, err)

	mag, err := preview.NewMagnet(inputMagnet)
	require.NoError(t, err)

	resolverRepo := new(clientmocks.MagnetClient)
	resolverRepo.On("Resolve", mock.Anything, mag).Return(torrent, nil)

	torrentRepo := new(storagemocks.TorrentRepository)
	torrentRepo.On("Persist", mock.Anything, torrent).Return(nil)
	torrentRepo.On("Get", mock.Anything, "cb84ccc10f296df72d6c40ba7a07c178a4323a14").
		Return(preview.Info{}, preview.ErrNotFound)

	eventBus := new(busmocks.Event)
	eventBus.On(
		"Publish",
		mock.Anything,
		&preview.TorrentCreatedEvent{TorrentID: "cb84ccc10f296df72d6c40ba7a07c178a4323a14"},
	).Return(nil)

	s := unmagnetize.NewService(fakeLogger(), eventBus, resolverRepo, torrentRepo)
	torrentReturned, err := s.Handle(context.Background(), unmagnetize.CMD{Magnet: inputMagnet})
	require.NoError(t, err)
	assert.Equal(t, torrent, torrentReturned)
}

func Test_MagnetService_Transform_AlreadyDownloaded(t *testing.T) {
	inputMagnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"
	torrentData := []byte("torrent-data")

	resolverRepo := new(clientmocks.MagnetClient)

	fakeTorrent, err := preview.NewInfo(
		"cb84ccc10f296df72d6c40ba7a07c178a4323a14",
		"test name",
		10,
		nil,
		nil,
	)
	require.NoError(t, err)

	torrentRepo := new(storagemocks.TorrentRepository)
	torrentRepo.On("Persist", mock.Anything, torrentData).Return(nil)
	torrentRepo.On("Get", mock.Anything, "cb84ccc10f296df72d6c40ba7a07c178a4323a14").Return(fakeTorrent, nil)

	eventBus := new(busmocks.Event)

	s := unmagnetize.NewService(fakeLogger(), eventBus, resolverRepo, torrentRepo)
	torrentID, err := s.Handle(context.Background(), unmagnetize.CMD{Magnet: inputMagnet})
	require.NoError(t, err)
	assert.Equal(t, fakeTorrent, torrentID)
}

func Test_MagnetService_Transform_RepositoryErrorGetTorrent(t *testing.T) {
	inputMagnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"
	torrentData := []byte("torrent-data")

	resolverRepo := new(clientmocks.MagnetClient)

	torrentRepo := new(storagemocks.TorrentRepository)
	torrentRepo.On("Persist", mock.Anything, torrentData).Return(nil)
	torrentRepo.On("Get", mock.Anything, "cb84ccc10f296df72d6c40ba7a07c178a4323a14").Return(preview.Info{}, errors.New("fake error"))

	eventBus := new(busmocks.Event)

	s := unmagnetize.NewService(fakeLogger(), eventBus, resolverRepo, torrentRepo)
	_, err := s.Handle(context.Background(), unmagnetize.CMD{Magnet: inputMagnet})
	require.Error(t, err)
}

func Test_MagnetService_Inspect_RepositoryError(t *testing.T) {
	inputMagnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"

	mag, err := preview.NewMagnet(inputMagnet)
	require.NoError(t, err)

	torrent, err := preview.NewInfo(
		"zocmzqipffw7ollmic5hub6bpcsdeoqu",
		"test torrent",
		100,
		nil,
		[]byte("torrent-data"),
	)
	require.NoError(t, err)

	resolverRepo := new(clientmocks.MagnetClient)
	resolverRepo.On("Resolve", mock.Anything, mag).
		Return(torrent, errors.New("fake repo error"))

	torrentRepo := new(storagemocks.TorrentRepository)
	torrentRepo.On("Get", mock.Anything, "cb84ccc10f296df72d6c40ba7a07c178a4323a14").
		Return(preview.Info{}, preview.ErrNotFound)

	eventBus := new(busmocks.Event)

	s := unmagnetize.NewService(fakeLogger(), eventBus, resolverRepo, torrentRepo)
	_, err = s.Handle(context.Background(), unmagnetize.CMD{Magnet: inputMagnet})
	require.Error(t, err)
}

func Test_MagnetService_Inspect_InvalidMagnetError(t *testing.T) {
	inputMagnet := "invalid magnet"

	resolverRepo := new(clientmocks.MagnetClient)
	torrentRepo := new(storagemocks.TorrentRepository)

	eventBus := new(busmocks.Event)

	s := unmagnetize.NewService(fakeLogger(), eventBus, resolverRepo, torrentRepo)
	_, err := s.Handle(context.Background(), unmagnetize.CMD{Magnet: inputMagnet})
	require.Error(t, err)
}

func fakeLogger() *logrus.Logger {
	l := logrus.New()
	l.Out = ioutil.Discard
	return l
}
