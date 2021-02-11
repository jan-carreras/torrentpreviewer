package transform_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/client/clientmocks"
	"prevtorrent/internal/preview/platform/storage/storagemocks"
	"prevtorrent/internal/preview/transform"
	"testing"
)

func Test_MagnetService_Inspect_Succeed(t *testing.T) {
	inputMagnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"
	torrentData := []byte("torrent-data")

	mag, err := preview.NewMagnet(inputMagnet)
	assert.NoError(t, err)

	resolverRepo := new(clientmocks.MagnetResolver)
	resolverRepo.On("Resolve", mock.Anything, mag).Return(torrentData, nil)

	torrentRepo := new(storagemocks.TorrentRepository)
	torrentRepo.On("Persist", mock.Anything, torrentData).Return(nil)

	s := transform.NewService(resolverRepo, torrentRepo)
	err = s.ToTorrent(context.Background(), transform.ServiceCMD{Magnet: inputMagnet})
	assert.NoError(t, err)
}

func Test_MagnetService_Inspect_RepositoryError(t *testing.T) {
	inputMagnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"

	mag, err := preview.NewMagnet(inputMagnet)
	assert.NoError(t, err)

	resolverRepo := new(clientmocks.MagnetResolver)
	resolverRepo.On("Resolve", mock.Anything, mag).Return(nil, errors.New("fake repo error"))

	torrentRepo := new(storagemocks.TorrentRepository)

	s := transform.NewService(resolverRepo, torrentRepo)
	err = s.ToTorrent(context.Background(), transform.ServiceCMD{Magnet: inputMagnet})
	assert.Error(t, err)
}

func Test_MagnetService_Inspect_InvalidMagnetError(t *testing.T) {
	inputMagnet := "invalid magnet"

	resolverRepo := new(clientmocks.MagnetResolver)
	torrentRepo := new(storagemocks.TorrentRepository)

	s := transform.NewService(resolverRepo, torrentRepo)
	err := s.ToTorrent(context.Background(), transform.ServiceCMD{Magnet: inputMagnet})
	assert.Error(t, err)
}
