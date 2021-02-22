package file

import (
	"context"
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

type ImagePersister struct {
	logger   *logrus.Logger
	imageDir string
}

func NewImagePersister(logger *logrus.Logger, imageDir string) *ImagePersister {
	return &ImagePersister{logger: logger, imageDir: imageDir}
}

func (r *ImagePersister) PersistFile(ctx context.Context, id string, data []byte) error {
	if err := ensureDirectoryExists(r.imageDir); err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(r.imageDir, id), data, 0644)
}

func ensureDirectoryExists(dir string) (err error) {
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		return nil
	}
	return err
}
