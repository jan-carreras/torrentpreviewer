package importTorrent_test

import (
	"context"
	"errors"
	"io/ioutil"
	"prevtorrent/internal/platform/bus/busmocks"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/importTorrent"
	"prevtorrent/internal/preview/platform/client/clientmocks"
	"prevtorrent/internal/preview/platform/storage/storagemocks"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Import_TorrentImportError(t *testing.T) {
	raw := []byte("")

	commandBus := new(busmocks.Command)
	torrentDownloader := new(clientmocks.TorrentDownloader)
	torrentDownloader.On("Import", mock.Anything, raw).
		Return(preview.Torrent{}, errors.New("fake error"))

	torrentRepository := new(storagemocks.TorrentRepository)

	service := importTorrent.NewService(fakeLogger(), commandBus, torrentDownloader, torrentRepository)

	_, err := service.Import(context.Background(), importTorrent.CMD{
		TorrentRaw: raw,
	})
	assert.Error(t, err)
}

func TestService_Import_TorrentAlreadyExists(t *testing.T) {
	raw := []byte("")

	fakeTorrent, err := preview.NewInfo("cb84ccc10f296df72d6c40ba7a07c178a4323a14", "test torrent", 100, nil, raw)
	require.NoError(t, err)

	commandBus := new(busmocks.Command)

	torrentDownloader := new(clientmocks.TorrentDownloader)
	torrentDownloader.On("Import", mock.Anything, raw).
		Return(fakeTorrent, nil)

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, "cb84ccc10f296df72d6c40ba7a07c178a4323a14").
		Return(preview.Torrent{}, nil)

	service := importTorrent.NewService(fakeLogger(), commandBus, torrentDownloader, torrentRepository)

	torrent, err := service.Import(context.Background(), importTorrent.CMD{
		TorrentRaw: raw,
	})
	assert.NoError(t, err)
	assert.Equal(t, fakeTorrent, torrent)
}

func TestService_Import_ErrorOnPersist(t *testing.T) {
	raw := []byte("")

	fakeTorrent, err := preview.NewInfo("cb84ccc10f296df72d6c40ba7a07c178a4323a14", "test torrent", 100, nil, raw)
	require.NoError(t, err)

	commandBus := new(busmocks.Command)

	torrentDownloader := new(clientmocks.TorrentDownloader)
	torrentDownloader.On("Import", mock.Anything, raw).
		Return(fakeTorrent, nil)

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, "cb84ccc10f296df72d6c40ba7a07c178a4323a14").
		Return(preview.Torrent{}, preview.ErrNotFound)
	torrentRepository.On("Persist", mock.Anything, fakeTorrent).
		Return(errors.New("fake error on persist"))

	service := importTorrent.NewService(fakeLogger(), commandBus, torrentDownloader, torrentRepository)

	_, err = service.Import(context.Background(), importTorrent.CMD{
		TorrentRaw: raw,
	})
	assert.Error(t, err)
}

func TestService_Import_ErrorOnEvent(t *testing.T) {
	raw := []byte("")

	fakeTorrent, err := preview.NewInfo("cb84ccc10f296df72d6c40ba7a07c178a4323a14", "test torrent", 100, nil, raw)
	require.NoError(t, err)

	commandBus := new(busmocks.Command)
	commandBus.On("Send", context.Background(), preview.NewTorrentCreatedEvent(fakeTorrent.ID())).
		Return(errors.New("fake error on send event"))

	torrentDownloader := new(clientmocks.TorrentDownloader)
	torrentDownloader.On("Import", mock.Anything, raw).
		Return(fakeTorrent, nil)

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, "cb84ccc10f296df72d6c40ba7a07c178a4323a14").
		Return(preview.Torrent{}, preview.ErrNotFound)
	torrentRepository.On("Persist", mock.Anything, fakeTorrent).
		Return(nil)

	service := importTorrent.NewService(fakeLogger(), commandBus, torrentDownloader, torrentRepository)

	_, err = service.Import(context.Background(), importTorrent.CMD{
		TorrentRaw: raw,
	})
	assert.Error(t, err)
}

func TestService_Import_BaseCase(t *testing.T) {
	raw := []byte("")

	fakeTorrent, err := preview.NewInfo("cb84ccc10f296df72d6c40ba7a07c178a4323a14", "test torrent", 100, nil, raw)
	require.NoError(t, err)

	commandBus := new(busmocks.Command)
	commandBus.On("Send", context.Background(), preview.NewTorrentCreatedEvent(fakeTorrent.ID())).
		Return(nil)

	torrentDownloader := new(clientmocks.TorrentDownloader)
	torrentDownloader.On("Import", mock.Anything, raw).
		Return(fakeTorrent, nil)

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, "cb84ccc10f296df72d6c40ba7a07c178a4323a14").
		Return(preview.Torrent{}, preview.ErrNotFound)
	torrentRepository.On("Persist", mock.Anything, fakeTorrent).
		Return(nil)

	service := importTorrent.NewService(fakeLogger(), commandBus, torrentDownloader, torrentRepository)

	torrent, err := service.Import(context.Background(), importTorrent.CMD{
		TorrentRaw: raw,
	})
	assert.NoError(t, err)
	assert.Equal(t, fakeTorrent, torrent)
}

func fakeLogger() *logrus.Logger {
	l := logrus.New()
	l.Out = ioutil.Discard
	return l
}
