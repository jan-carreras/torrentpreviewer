package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
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
	frameExtractionTime := "0:00:03.000"

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	filename := path.Join(os.TempDir(), fmt.Sprintf("prevtorrent.ffmpgout.%v.jpg", id))
	defer rmImage(filename)

	cmd := exec.Command(command,
		"-ss", frameExtractionTime,
		"-i", readStdin,
		"-vframes", vframes,
		"-q:v", qv,
		filename,
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
	return ioutil.ReadFile(filename)
}

func rmImage(src string) {
	_ = os.Remove(src)
}
