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

func (r *TorrentRepository) Get(ctx context.Context, id string) (preview.Info, error) {
	raw, err := ioutil.ReadFile(id)
	if err != nil {
		return preview.Info{}, err
	}

	i, err := r.parseMetaInfoFromRaw(raw)
	if err != nil {
		return preview.Info{}, err
	}

	files, err := r.parseFileInfo(i)
	if err != nil {
		return preview.Info{}, err
	}

	return preview.NewInfo(
		i.Name,
		int(i.PieceLength),
		i.NumPieces(),
		files,
		raw,
	)
}

func (r *TorrentRepository) parseMetaInfoFromRaw(raw []byte) (metainfo.Info, error) {
	metaInfo := new(metainfo.MetaInfo)
	err := bencode.Unmarshal(raw, metaInfo)
	if err != nil {
		return metainfo.Info{}, err
	}

	i, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return metainfo.Info{}, nil
	}
	return i, nil
}

func (r *TorrentRepository) parseFileInfo(i metainfo.Info) ([]preview.FileInfo, error) {
	files := make([]preview.FileInfo, 0)
	for _, file := range i.UpvertedFiles() {
		if len(file.Path) == 0 {
			// TODO logger.Printf("torrent has 0 files. should at least have one")
			continue
		}
		if len(file.Path) > 1 {
			// TODO: Add proper logging
		}

		fi, err := preview.NewFileInfo(
			int(file.Length),
			file.Path[0],
			file.DisplayPath(&i),
		)
		if err != nil {
			return nil, err
		}
		files = append(files, fi)
	}
	return files, nil
}
