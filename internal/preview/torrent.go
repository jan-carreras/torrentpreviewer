package preview

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	mb           = 1 << (10 * 2) // MiB, really
	DownloadSize = 8 * mb
)

var torrentIDValidation = regexp.MustCompile("^([a-zA-Z0-9]+)$")

//go:generate mockery --case=snake --outpkg=storagemocks --output=platform/storage/storagemocks --name=TorrentRepository
type TorrentRepository interface {
	Persist(ctx context.Context, torrent Torrent) error
	Get(ctx context.Context, id string) (Torrent, error)
}

//go:generate mockery --case=snake --outpkg=clientmocks --output=platform/client/clientmocks --name=TorrentDownloader
type TorrentDownloader interface {
	DownloadParts(context.Context, DownloadPlan) (*PieceRegistry, error)
	Import(ctx context.Context, raw []byte) (Torrent, error)
}

var ErrNotFound = errors.New("record not found in storage")

// Torrent represents the torrent meta info.
// It's not a 1-1 map of a torrent file, just what we need to download some parts
type Torrent struct {
	id          string // id of the torrent
	name        string // name of the torrent. might be empty
	pieceLength int    // pieceLength for the whole torrent
	totalLength int    // totalLength is the sum in bytes of all files
	files       []File // files is a list of all the files present in the torrent. should never be empty
	filesByID   map[int]*File
	raw         []byte // raw is the bencoded representation of a .torrent
}

var ErrInvalidTorrentID = errors.New("invalid torrent ID")

// NewInfo create a new torrent
func NewInfo(
	id string,
	name string,
	pieceLength int,
	files []File,
	raw []byte,
) (Torrent, error) {
	if !torrentIDValidation.MatchString(id) {
		return Torrent{}, ErrInvalidTorrentID
	}

	id, err := toBase32(id)
	if err != nil {
		return Torrent{}, err
	}
	if len(id) != 40 {
		return Torrent{}, errors.New("id must have 32 chars (hex encoded) or 40 chars (base32 encoded)")
	}

	if err := validateFiles(files); err != nil {
		return Torrent{}, err
	}

	return Torrent{
		id:          strings.ToLower(id),
		name:        name,
		pieceLength: pieceLength,
		totalLength: totalLength(files),
		files:       files,
		raw:         raw,
		filesByID:   filesByID(files),
	}, nil
}

// ID returns the torrent ID
func (i Torrent) ID() string {
	return i.id
}

// Raw returns the bencoded representation of a torrent
func (i Torrent) Raw() []byte {
	return i.raw
}

// Name returns the name of the torrent. Might be empty
func (i Torrent) Name() string {
	return i.name
}

// TotalLength return the sum of all files in bytes
func (i Torrent) TotalLength() int {
	return i.totalLength
}

// PieceLength returns the size of each piece
func (i Torrent) PieceLength() int {
	return i.pieceLength
}

// Files is a list of files that this torrent holds. Should not be empty
func (i Torrent) Files() []File {
	return i.files
}

func (i Torrent) File(idx int) *File {
	return i.filesByID[idx]
}

// SupportedFiles returns from all the files, the ones that have an extension supported by ffmpeg
func (i Torrent) SupportedFiles() []File {
	fi := make([]File, 0)
	for _, f := range i.files {
		_f := f
		if f.IsSupportedExtension() {
			fi = append(fi, _f)
		}
	}
	return fi
}

// File describes each file on the torrent.
// Each file is identified by its position (which is important), the length and an arbitrary name
type File struct {
	idx    int
	length int
	name   string
	images []Image
}

// NewFileInfo creates a File
func NewFileInfo(idx int, length int, name string) (File, error) {
	return File{idx: idx, length: length, name: name, images: make([]Image, 0)}, nil
}

func (fi File) IsEqual(fi2 File) bool {
	return fi.idx == fi2.idx
}

// ID returns the ID, which is the index on the list of files of the torrent. zero indexed.
func (fi File) ID() int {
	return fi.idx
}

// Length returns the length of the file
func (fi File) Length() int {
	return fi.length
}

// Name returns the name of the file
func (fi File) Name() string {
	return fi.name
}

// DownloadSize is how much are we going to download from the file.
// Either a fixed amount or the whole file is smaller
func (fi File) DownloadSize() int {
	size := DownloadSize
	if size > fi.length {
		return fi.length
	}
	return DownloadSize
}

// IsSupportedExtension returns is the file has a supported extension to generate a preview
func (fi File) IsSupportedExtension() bool {
	supported := map[string]interface{}{
		".mp4": struct{}{},
		".mkv": struct{}{},
		".mov": struct{}{},
	}

	ext := filepath.Ext(fi.name)
	_, found := supported[ext]
	return found
}

func (fi *File) AddImage(image Image) error {
	if image.fileID != fi.ID() {
		return fmt.Errorf("the image with name '%v' and fileID '%v' does not match fileID %v ",
			image.Name(),
			image.fileID,
			fi.ID(),
		)
	}
	fi.images = append(fi.images, image)
	return nil
}

func (fi File) Images() []Image {
	return fi.images
}
