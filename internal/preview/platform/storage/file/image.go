package file

import (
	"context"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
)

type ImageRepository struct {
	logger   *logrus.Logger
	imageDir string
}

func NewImageRepository(logger *logrus.Logger, imageDir string) *ImageRepository {
	return &ImageRepository{logger: logger, imageDir: imageDir}
}

func (r *ImageRepository) PersistFile(ctx context.Context, id string, data []byte) error {
	if err := ensureDirectoryExists(r.imageDir); err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(r.imageDir, id), data, 0600)
}

func ensureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	return nil
}
