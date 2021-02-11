package http

import (
	"bytes"
	"context"
	"errors"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"io/ioutil"
	"path"
	"prevtorrent/internal/preview"
)

type TorrentIntegration struct {
	client *torrent.Client
}

func NewTorrentIntegration(client *torrent.Client) *TorrentIntegration {
	return &TorrentIntegration{client: client}
}

func (r *TorrentIntegration) Resolve(ctx context.Context, m preview.Magnet) ([]byte, error) {
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

func (r *TorrentIntegration) waitForInfo(ctx context.Context, t *torrent.Torrent) error {
	select {
	case <-t.GotInfo():
		return nil
	case <-ctx.Done():
		return errors.New("context cancelled while trying to get info")
	}
}

type TorrentRepository struct {
	torrentDir string
}

func NewTorrentRepository() *TorrentRepository {
	return &TorrentRepository{}
}

func (r *TorrentRepository) Persist(ctx context.Context, data []byte) error {
	buf := bytes.NewBuffer(data)
	d := bencode.NewDecoder(buf)

	metaInfo := new(metainfo.MetaInfo)
	err := d.Decode(metaInfo)
	if err != nil {
		return err
	}

	info, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join("/tmp/", info.Name+".torrent"), data, 0666)
}
