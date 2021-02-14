package preview

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
)

const (
	mb           = 1 << (10 * 2) // MiB, really
	downloadSize = 8 * mb
)

//go:generate mockery --case=snake --outpkg=storagemocks --output=platform/storage/storagemocks --name=TorrentRepository
type TorrentRepository interface {
	Persist(ctx context.Context, torrent Info) error
	Get(ctx context.Context, id string) (Info, error)
}

var ErrNotFound = errors.New("record not found in storage")

// TODO: That's the worst name of the whole universe
type Info struct {
	id          string
	name        string
	pieceLength int
	totalLength int
	pieces      int
	files       []FileInfo
	raw         []byte
}

var ErrInfoNameCannotBeEmpty = errors.New("info.name cannot be empty")

func NewInfo(
	id string,
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
		id:          strings.ToLower(id),
		name:        name,
		pieceLength: pieceLength,
		totalLength: 0, // TODO: Fill this
		pieces:      pieces,
		files:       files,
		raw:         raw,
	}, nil
}

func (i Info) ID() string {
	return i.id
}

func (i Info) Raw() []byte {
	return i.raw
}

func (i Info) Name() string {
	return i.name
}

func (i Info) TotalLength() int {
	return i.totalLength
}

func (i Info) PieceLength() int {
	return i.pieceLength
}

func (i Info) Files() []FileInfo {
	return i.files
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
}

func NewFileInfo(idx int, length int, name string) (FileInfo, error) {
	return FileInfo{idx: idx, length: length, name: name}, nil
}

func (fi FileInfo) ID() int {
	return fi.idx
}

func (fi FileInfo) Length() int {
	return fi.length
}

func (fi FileInfo) Name() string {
	return fi.name
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
