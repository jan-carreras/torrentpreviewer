package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"prevtorrent/internal/preview"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	command = "ffmpeg"
	vframes = "1"
	qv      = "2"
)

type InMemoryFfmpeg struct {
	logger *logrus.Logger
}

func NewInMemoryFfmpeg(logger *logrus.Logger) (*InMemoryFfmpeg, error) {
	if err := checkFFMPGExecutableIsInPath(); err != nil {
		return nil, err
	}
	return &InMemoryFfmpeg{
		logger: logger,
	}, nil
}

func checkFFMPGExecutableIsInPath() error {
	cmd := exec.Command(command, "-version")
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (i *InMemoryFfmpeg) ExtractImage(ctx context.Context, data []byte, time int) ([]byte, error) {
	frameExtractionTime := strconv.Itoa(time)

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	outputFilename := path.Join(os.TempDir(), fmt.Sprintf("prevtorrent.ffmpgout.%v.jpg", id.String()))
	defer rmFile(outputFilename)

	// IMPROVEMENT: For some reason passing the file from STDIN (see below) crashes ffmpeg.
	//       Doing it with a file seems to work better but involves IO. Would be nice to get rid of it
	//       in the future or use tmpfs instead
	inputVideo := outputFilename + ".mp4"
	err = ioutil.WriteFile(inputVideo, data, 0600)
	if err != nil {
		return nil, err
	}
	defer rmFile(inputVideo)

	cmd := exec.Command(command,
		"-ss", frameExtractionTime, // Always keep before the -i option for performance considerations! https://trac.ffmpeg.org/wiki/Seeking
		"-i", inputVideo,
		"-vframes", vframes,
		"-q:v", qv,
		outputFilename,
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

	img, err := ioutil.ReadFile(outputFilename)
	if err != nil {
		return nil, i.logCommandFailed(err, stdOut, stdErr)
	}

	return img, err
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

	if os.IsNotExist(err) {
		err = fmt.Errorf("%w. %v", preview.ErrNotAbleToGenerateImage, err)
	}

	return err
}

func (i *InMemoryFfmpeg) isAtomNotFound(stderr string) bool {
	return strings.Contains(stderr, "moov atom not found")
}

func rmFile(src string) {
	_ = os.Remove(src)
}
