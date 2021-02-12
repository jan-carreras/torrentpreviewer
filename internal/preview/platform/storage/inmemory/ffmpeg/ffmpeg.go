package ffmpeg

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
)

const (
	command   = "ffmpeg"
	vframes   = "1"
	qv        = "2"
	readStdin = "-"
)

type InMemoryFfmpeg struct {
	logger *logrus.Logger
}

func NewInMemoryFfmpeg(logger *logrus.Logger) *InMemoryFfmpeg {
	return &InMemoryFfmpeg{
		logger: logger,
	}
}

func (i *InMemoryFfmpeg) ExtractImage(ctx context.Context, data []byte, time int) ([]byte, error) {
	// TODO: Simplify the randomness

	frameExtractionTime := "0:00:05.000"

	f, err := ioutil.TempFile("/tmp", "image.*.jpg")
	if err != nil {
		return nil, err
	}
	f.Close()
	tryDeleteImage(f.Name())       // ffmpeg needs the file to not exist
	defer tryDeleteImage(f.Name()) // ffmpeg will create the image again. We need it gone

	cmd := exec.Command(command,
		"-ss", frameExtractionTime,
		"-i", readStdin,
		"-vframes", vframes,
		"-q:v", qv,
		f.Name(),
	)

	cmd.Stdin = bytes.NewBuffer(data)
	stdOut := new(bytes.Buffer)
	cmd.Stdout = stdOut
	stdErr := new(bytes.Buffer)
	cmd.Stderr = stdErr

	if err := cmd.Start(); err != nil {
		i.logger.WithFields(logrus.Fields{
			"stdout": stdOut.String(),
			"stderr": stdErr.String(),
			"err":    err.Error(),
		}).Warn("command failed")
		return nil, errors.Wrap(err, "error while executing ffmpeg  the command")
	}
	if err := cmd.Wait(); err != nil {
		i.logger.WithFields(logrus.Fields{
			"stdout": stdOut.String(),
			"stderr": stdErr.String(),
			"err":    err.Error(),
		}).Warn("command failed")

		return nil, errors.Wrap(err, "error while waiting for ffmpeg command to finish")
	}
	return ioutil.ReadFile(f.Name())
}

func tryDeleteImage(src string) {
	_ = os.Remove(src)
}
