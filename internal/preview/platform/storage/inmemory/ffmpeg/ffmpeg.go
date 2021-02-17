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
	"prevtorrent/internal/preview"
	"strings"
)

const (
	command = "ffmpeg"
	vframes = "1"
	qv      = "2"
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
	frameExtractionTime := "0:00:05.000" // TODO: Use time parameter instead

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	filename := path.Join(os.TempDir(), fmt.Sprintf("prevtorrent.ffmpgout.%v.jpg", id.String()))
	defer rmFile(filename)

	// TODO: For some reason passing the file from STDIN (see below) crashes ffmpeg.
	//       Doing it with a file seems to work better but involves IO. Would be nice to get rid of it
	//       in the future or use tmpfs instead
	video := filename + ".mp4"
	err = ioutil.WriteFile(video, data, 0600)
	if err != nil {
		return nil, err
	}
	defer rmFile(video)

	cmd := exec.Command(command,
		"-ss", frameExtractionTime,
		"-i", video,
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
		err = errors.Wrap(err, "error while executing ffmpeg  the command")
		return nil, i.logCommandFailed(err, stdOut, stdErr)
	}
	if err := cmd.Wait(); err != nil {
		err = errors.Wrap(err, "error while waiting for ffmpeg command to finish")
		return nil, i.logCommandFailed(err, stdOut, stdErr)
	}
	return ioutil.ReadFile(filename)
}

func (i *InMemoryFfmpeg) logCommandFailed(err error, stdOut, stdErr *bytes.Buffer) error {
	i.logger.WithFields(logrus.Fields{
		"stdout": stdOut.String(),
		"stderr": stdErr.String(),
		"err":    err.Error(),
	}).Warn("command failed")

	if i.isAtomNotFound(stdErr.String()) {
		err = fmt.Errorf("%w. %v", preview.ErrAtomNotFound, err)
	}

	return err
}

func (i *InMemoryFfmpeg) isAtomNotFound(stderr string) bool {
	return strings.Index(stderr, "moov atom not found") >= 0
}

func rmFile(src string) {
	_ = os.Remove(src)
}
