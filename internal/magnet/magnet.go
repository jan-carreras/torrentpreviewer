package magnet

import (
	"context"
	"errors"
	"regexp"
)

var magnetValidationRegexp = regexp.MustCompile("magnet:\\?xt=urn:btih:[a-zA-Z0-9]*")

//go:generate mockery --case=snake --outpkg=clientmocks --output=platform/client/clientmocks --name=MagnetResolver
type MagnetResolver interface {
	Resolve(context.Context, Magnet) ([]byte, error)
}

//go:generate mockery --case=snake --outpkg=clientmocks --output=platform/client/clientmocks --name=TorrentRepository
type TorrentRepository interface {
	Persist(context.Context, []byte) error
}

type Magnet struct {
	value string
}

var ErrInvalidMagnetFormat = errors.New("invalid magnet")

func NewMagnet(value string) (Magnet, error) {
	if !magnetValidationRegexp.Match([]byte(value)) {
		return Magnet{}, ErrInvalidMagnetFormat
	}
	return Magnet{value: value}, nil
}

func (m Magnet) Value() string {
	return m.value
}

type Info struct {
	pieceLength int
	name        string
	length      int // mutually exclusive with Files
	files       []FileInfo
}

var ErrInfoNameCannotBeEmpty = errors.New("info.name cannot be empty")
var ErrLengthAndFilesAreMutuallyExclusive = errors.New("length and files are mutually exclusive")

func NewInfo(
	pieceLength int,
	name string,
	length int,
	files []FileInfo,
) (Info, error) {
	if name == "" {
		return Info{}, ErrInfoNameCannotBeEmpty
	}
	if length != 0 && len(files) != 0 {
		return Info{}, ErrLengthAndFilesAreMutuallyExclusive
	}

	return Info{pieceLength: pieceLength, name: name, length: length, files: files}, nil
}

type FileInfo struct {
	length int
	path   []string
}

func NewFileInfo(l int, p []string) (FileInfo, error) {
	return FileInfo{length: l, path: p}, nil
}
