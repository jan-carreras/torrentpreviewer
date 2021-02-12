package file

import (
	"bytes"
	"context"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"prevtorrent/internal/preview"
)

type TorrentRepository struct {
	torrentDir string
	logger     *logrus.Logger
}

func NewTorrentRepository(
	torrentDir string,
	logger *logrus.Logger,
) *TorrentRepository {
	return &TorrentRepository{
		torrentDir: torrentDir,
		logger:     logger,
	}
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

	if err := r.ensureDirectoryExists(); err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(r.torrentDir, info.Name+".torrent"), data, 0600)
}

func (r *TorrentRepository) ensureDirectoryExists() error {
	if _, err := os.Stat(r.torrentDir); os.IsNotExist(err) {
		if err := os.MkdirAll(r.torrentDir, 0700); err != nil {
			return err
		}
	}
	return nil
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
		filePath := file.DisplayPath(&i)
		if len(filePath) == 0 {
			r.logger.WithFields(logrus.Fields{
				"torrent":  i.Name,
				"filePath": filePath,
			}).Warn("torrent filename is empty. shouldn't be something? ignoring")
			continue
		}

		fi, err := preview.NewFileInfo(
			int(file.Length),
			filePath,
			file.DisplayPath(&i),
		)
		if err != nil {
			return nil, err
		}
		files = append(files, fi)
	}
	return files, nil
}
