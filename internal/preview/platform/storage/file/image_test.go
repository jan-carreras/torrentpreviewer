package file_test

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"prevtorrent/internal/preview/platform/storage/file"
	"testing"
)

func TestImagePersister_PersistFile(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "torrentpreviewtest")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	imageName := "image.jpg"
	imageData := []byte("example")

	imagePersister := file.NewImagePersister(fakeLogger(), dir)
	err = imagePersister.PersistFile(context.Background(), imageName, imageData)
	assert.NoError(t, err)

	newImagePath := path.Join(dir, "image.jpg")

	content, err := ioutil.ReadFile(newImagePath)
	assert.NoError(t, err)
	assert.Equal(t, imageData, content)

	fi, err := os.Stat(newImagePath)
	assert.NoError(t, err)
	assert.Equal(t, "-rw-r--r--", fi.Mode().String())
}

func TestImagePersister_PersistFile_CreatesDirectoryIfNotExists(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "torrentpreviewtest")
	assert.NoError(t, err)
	err = os.RemoveAll(dir)
	defer os.RemoveAll(dir)
	assert.NoError(t, err)

	imagePersister := file.NewImagePersister(fakeLogger(), dir)
	err = imagePersister.PersistFile(context.Background(), "image.jpg", []byte("example"))
	assert.NoError(t, err)
}

func TestImagePersister_PersistFile_ErrIfUnableToCreateFile(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "torrentpreviewtest")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	assert.NoError(t, os.Chmod(dir, 0000))

	imagePersister := file.NewImagePersister(fakeLogger(), dir)
	err = imagePersister.PersistFile(context.Background(), "image.jpg", []byte("example"))
	assert.Error(t, err)
}

func fakeLogger() *logrus.Logger {
	l := logrus.New()
	l.Out = ioutil.Discard
	return l
}
