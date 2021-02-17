package file

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"prevtorrent/internal/preview"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/sirupsen/logrus"
)

type FileTorrentRepository struct {
	torrentDir   string
	downloadsDir string
	logger       *logrus.Logger
}

func NewTorrentRepository(
	logger *logrus.Logger,
	torrentDir string,
	downloadsDir string,
) *FileTorrentRepository {
	return &FileTorrentRepository{
		logger:       logger,
		torrentDir:   torrentDir,
		downloadsDir: downloadsDir,
	}
}

func (r *FileTorrentRepository) Persist(ctx context.Context, data []byte) error {
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

	if err := ensureDirectoryExists(r.torrentDir); err != nil {
		return err
	}
	dst := path.Join(r.torrentDir, info.Name+".torrent")
	return ioutil.WriteFile(dst, data, 0600)
}

func (r *FileTorrentRepository) Get(ctx context.Context, id string) (preview.Info, error) {
	raw, err := ioutil.ReadFile(id)
	if err != nil {
		return preview.Info{}, fmt.Errorf("%w. %v", preview.ErrNotFound, err)
	}

	torrentID, i, err := r.parseMetaInfoFromRaw(raw)
	if err != nil {
		return preview.Info{}, err
	}

	files, err := r.parseFileInfo(i)
	if err != nil {
		return preview.Info{}, err
	}

	return preview.NewInfo(
		torrentID,
		i.Name,
		int(i.PieceLength),
		files,
		raw,
	)
}

func (r *FileTorrentRepository) parseMetaInfoFromRaw(raw []byte) (string, metainfo.Info, error) {
	metaInfo := new(metainfo.MetaInfo)
	err := bencode.Unmarshal(raw, metaInfo)
	if err != nil {
		return "", metainfo.Info{}, err
	}

	i, err := metaInfo.UnmarshalInfo()
	if err != nil {
		return "", metainfo.Info{}, nil
	}
	id := metaInfo.HashInfoBytes().HexString()
	return id, i, nil
}

func (r *FileTorrentRepository) parseFileInfo(i metainfo.Info) ([]preview.FileInfo, error) {
	files := make([]preview.FileInfo, 0)
	for idx, file := range i.UpvertedFiles() {
		filePath := file.DisplayPath(&i)
		if len(filePath) == 0 {
			r.logger.WithFields(logrus.Fields{
				"torrent":  i.Name,
				"filePath": filePath,
			}).Warn("torrent filename is empty. shouldn't be something? ignoring")
			continue
		}

		fi, err := preview.NewFileInfo(
			idx,
			int(file.Length),
			file.DisplayPath(&i),
		)
		if err != nil {
			return nil, err
		}
		files = append(files, fi)
	}
	return files, nil
}
