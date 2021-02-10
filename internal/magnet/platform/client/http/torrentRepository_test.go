package http_test

import (
	"context"
	"errors"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	magnet2 "prevtorrent/internal/magnet"
	"prevtorrent/internal/magnet/platform/client/http"
	"prevtorrent/internal/magnet/platform/client/http/httpmocks"
	"testing"
)

func TestTorrentRepository_DescribeFiles_Success(t *testing.T) {
	magnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"
	pieceLength := 100
	length := 1000
	name := "magnet-name"

	m, err := magnet2.NewMagnet(magnet)
	assert.NoError(t, err)

	expecingInfo, err := magnet2.NewInfo(pieceLength, name, length, nil)
	assert.NoError(t, err)

	externalInfo := &metainfo.Info{
		PieceLength: int64(pieceLength),
		Name:        name,
		Length:      int64(length),
		Files:       []metainfo.FileInfo{},
	}

	client := new(httpmocks.TorrentClient)
	client.On("GetInfo", mock.Anything, magnet).Return(externalInfo, nil)

	repo := http.NewTorrentRepository(client)
	info, err := repo.GetMagnetInfo(context.Background(), m)
	assert.NoError(t, err)
	assert.Equal(t, expecingInfo, info)
}

func TestTorrentRepository_DescribeFiles_ErrorOnGetInfo(t *testing.T) {
	magnet := "magnet:?xt=urn:btih:ZOCMZQIPFFW7OLLMIC5HUB6BPCSDEOQU"

	m, err := magnet2.NewMagnet(magnet)
	assert.NoError(t, err)

	externalInfo := &metainfo.Info{}

	client := new(httpmocks.TorrentClient)
	client.On("GetInfo", mock.Anything, magnet).Return(externalInfo, errors.New("fake getInfo error"))

	repo := http.NewTorrentRepository(client)
	_, err = repo.GetMagnetInfo(context.Background(), m)
	assert.Error(t, err)
}
