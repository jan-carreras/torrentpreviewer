package preview

import "context"

//go:generate mockery --case=snake --outpkg=storagemocks --output=platform/storage/storagemocks --name=ImageExtractor
type ImageExtractor interface {
	ExtractImage(ctx context.Context, data []byte, time int) ([]byte, error)
}

//go:generate mockery --case=snake --outpkg=storagemocks --output=platform/storage/storagemocks --name=ImageRepository
type ImageRepository interface {
	PersistFile(ctx context.Context, id string, data []byte) error
}
