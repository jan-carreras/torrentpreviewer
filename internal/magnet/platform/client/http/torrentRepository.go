package http

import (
	"context"
	"errors"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"prevtorrent/internal/magnet"
)

//go:generate mockery --case=snake --outpkg=httpmocks --output=httpmocks --name=TorrentClient
type TorrentClient interface {
	GetInfo(ctx context.Context, magnet string) (*metainfo.Info, error)
}

type TorrentRepository struct {
	client TorrentClient
}

func NewTorrentRepository(client TorrentClient) *TorrentRepository {
	return &TorrentRepository{client: client}
}

func (r *TorrentRepository) GetMagnetInfo(ctx context.Context, m magnet.Magnet) (magnet.Info, error) {
	info, err := r.client.GetInfo(ctx, m.Value())
	if err != nil {
		return magnet.Info{}, err
	}

	return magnet.NewInfo(int(info.PieceLength), info.Name, int(info.Length), nil)
}

type TorrentIntegration struct {
	client *torrent.Client
}

func NewTorrentIntegration(client *torrent.Client) *TorrentIntegration {
	return &TorrentIntegration{client: client}
}

func (i *TorrentIntegration) GetInfo(ctx context.Context, magnet string) (*metainfo.Info, error) {
	t, err := i.client.AddMagnet(magnet)
	if err != nil {
		return nil, err
	}

	select {
	case <-t.GotInfo():
		return t.Info(), nil
	case <-ctx.Done():
		return nil, errors.New("context cancelled while trying to get info")
	}
}
