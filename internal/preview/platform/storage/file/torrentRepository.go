package file

import (
	"bytes"
	"context"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"io/ioutil"
	"path"
	"prevtorrent/internal/preview"
)

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

func (r *TorrentRepository) Get(ctx context.Context, name string) (preview.Info, error) {
	return preview.Info{}, nil
}
