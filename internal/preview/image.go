package preview

import (
	"context"
	"errors"
)

var ErrAtomNotFound = errors.New("moov atom not found")

//go:generate mockery --case=snake --outpkg=storagemocks --output=platform/storage/storagemocks --name=ImageExtractor
type ImageExtractor interface {
	ExtractImage(ctx context.Context, data []byte, time int) ([]byte, error)
}

//go:generate mockery --case=snake --outpkg=storagemocks --output=platform/storage/storagemocks --name=ImagePersister
type ImagePersister interface {
	PersistFile(ctx context.Context, id string, data []byte) error
}

//go:generate mockery --case=snake --outpkg=storagemocks --output=platform/storage/storagemocks --name=ImageRepository
type ImageRepository interface {
	ByTorrent(ctx context.Context, id string) (*TorrentImages, error)
	Persist(ctx context.Context, img Image) error
}

type Image struct {
	torrentID string
	fileID    int
	name      string
	length    int
}

func NewImage(torrentID string, fileID int, name string, length int) Image {
	return Image{torrentID: torrentID, fileID: fileID, name: name, length: length}
}

func (i Image) TorrentID() string {
	return i.torrentID
}

func (i Image) FileID() int {
	return i.fileID
}

func (i Image) Name() string {
	return i.name
}

func (i Image) Length() int {
	return i.length
}
