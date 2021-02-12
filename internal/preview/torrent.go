package preview

import (
	"context"
	"errors"
	"path/filepath"
)

const (
	mb           = 1 << (10 * 2) // MiB, really
	downloadSize = 8 * mb
)

//go:generate mockery --case=snake --outpkg=storagemocks --output=platform/storage/storagemocks --name=TorrentRepository
type TorrentRepository interface {
	Persist(ctx context.Context, data []byte) error
	Get(ctx context.Context, id string) (Info, error)
	PersistFile(ctx context.Context, id string, data []byte) error
}

type Info struct {
	name        string
	pieceLength int
	pieces      int
	files       []FileInfo
	raw         []byte
}

func (i Info) Raw() []byte {
	return i.raw
}

var ErrInfoNameCannotBeEmpty = errors.New("info.name cannot be empty")

func NewInfo(
	name string,
	pieceLength int,
	pieces int,
	files []FileInfo,
	raw []byte,
) (Info, error) {
	if name == "" {
		return Info{}, ErrInfoNameCannotBeEmpty
	}
	return Info{
		name:        name,
		pieceLength: pieceLength,
		pieces:      pieces,
		files:       files,
		raw:         raw,
	}, nil
}

func (i Info) SupportedFiles() []FileInfo {
	fi := make([]FileInfo, 0)
	for _, f := range i.files {
		_f := f
		if f.IsSupportedExtension() {
			fi = append(fi, _f)
		}
	}
	return fi
}

type FileInfo struct {
	idx    int
	length int
	name   string
	path   string
}

func (fi FileInfo) DownloadSize() int {
	size := downloadSize
	if size > fi.length {
		return fi.length
	}
	return downloadSize
}

func (fi FileInfo) IsSupportedExtension() bool {
	supported := map[string]interface{}{
		".mp4": struct{}{},
	}

	ext := filepath.Ext(fi.name)
	_, found := supported[ext]
	return found
}

func NewFileInfo(idx int, length int, name string, path string) (FileInfo, error) {
	return FileInfo{idx: idx, length: length, name: name, path: path}, nil
}
