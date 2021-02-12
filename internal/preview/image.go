package preview

import "context"

// TODO: Check out how we are supposed to store those mocks. Put them where they make sense
//go:generate mockery --case=snake --outpkg=inmemorymocks --output=platform/storage/inmemory/inmemorymocks --name=ImageExtractor
type ImageExtractor interface {
	ExtractImage(ctx context.Context, data []byte, time int) ([]byte, error)
}

//go:generate mockery --case=snake --outpkg=filemocks --output=platform/storage/file/filemocks --name=ImageRepository
type ImageRepository interface {
	PersistFile(ctx context.Context, id string, data []byte) error
}
