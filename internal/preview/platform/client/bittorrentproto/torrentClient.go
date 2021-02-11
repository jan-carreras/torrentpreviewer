package bittorrentproto

import (
	"bytes"
	"context"
	"errors"
	"github.com/anacrolix/torrent"
	"prevtorrent/internal/preview"
)

type TorrentClient struct {
	client *torrent.Client
}

func NewTorrentClient(client *torrent.Client) *TorrentClient {
	return &TorrentClient{client: client}
}

func (r *TorrentClient) Resolve(ctx context.Context, m preview.Magnet) ([]byte, error) {
	t, err := r.client.AddMagnet(m.Value())
	if err != nil {
		return nil, err
	}

	if err := r.waitForInfo(ctx, t); err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = t.Metainfo().Write(buf)
	return buf.Bytes(), err
}

func (r *TorrentClient) waitForInfo(ctx context.Context, t *torrent.Torrent) error {
	select {
	case <-t.GotInfo():
		return nil
	case <-ctx.Done():
		return errors.New("context cancelled while trying to get info")
	}
}
