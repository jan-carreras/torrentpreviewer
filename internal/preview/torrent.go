package preview

import (
	"context"
	"errors"
)

//go:generate mockery --case=snake --outpkg=storagemocks --output=platform/storage/storagemocks --name=TorrentRepository
type TorrentRepository interface {
	Persist(ctx context.Context, data []byte) error
	Get(ctx context.Context, name string) (Info, error)
}

type Info struct {
	pieceLength int
	pieces      int
	name        string
	files       []FileInfo
	raw         []byte
}

var ErrInfoNameCannotBeEmpty = errors.New("info.name cannot be empty")

func NewInfo(
	pieceLength int,
	name string,
	files []FileInfo,
) (Info, error) {
	if name == "" {
		return Info{}, ErrInfoNameCannotBeEmpty
	}

	return Info{pieceLength: pieceLength, name: name, files: files}, nil
}

type FileInfo struct {
	length int
	name   string
	path   string
}

func NewFileInfo(l int, p string) (FileInfo, error) {
	return FileInfo{length: l, path: p}, nil
}
