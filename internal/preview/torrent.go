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
	Persist(ctx context.Context, torrent Info) error
	Get(ctx context.Context, id string) (Info, error)
}

//go:generate mockery --case=snake --outpkg=clientmocks --output=platform/client/clientmocks --name=TorrentDownloader
type TorrentDownloader interface {
	DownloadParts(context.Context, DownloadPlan) (*PieceRegistry, error)
}

var ErrNotFound = errors.New("record not found in storage")

// Info represents a Torrent information. For a torrent-related library it might be the **worst**
// name of all times. I feel ashamed.
// It's not a 1-1 map of a torrent file, just what we need to download some parts
type Info struct {
	id          string     // id of the torrent
	name        string     // name of the torrent. might be empty
	pieceLength int        // pieceLength for the whole torrent
	totalLength int        // totalLength is the sum in bytes of all files
	files       []FileInfo // files is a list of all the files present in the torrent. should never be empty
	filesByID   map[int]*FileInfo
	raw         []byte // raw is the bencoded representation of a .torrent
}

var ErrInvalidTorrentID = errors.New("invalid torrent ID")

// NewInfo create a new torrent
func NewInfo(
	id string,
	name string,
	pieceLength int,
	files []FileInfo,
	raw []byte,
) (Info, error) {
	if !torrentIDValidation.MatchString(id) {
		return Info{}, ErrInvalidTorrentID
	}

	totalLength := func(files []FileInfo) (length int) {
		for _, f := range files {
			length += f.Length()
		}
		return length
	}
	filesByID := func(files []FileInfo) map[int]*FileInfo {
		hash := make(map[int]*FileInfo)
		for i := 0; i < len(files); i++ {
			hash[files[i].ID()] = &files[i]
		}
		return hash
	}

	return Info{
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
func (i Info) ID() string {
	return i.id
}

// Raw returns the bencoded representation of a torrent
func (i Info) Raw() []byte {
	return i.raw
}

// Name returns the name of the torrent. Might be empty
func (i Info) Name() string {
	return i.name
}

// TotalLength return the sum of all files in bytes
func (i Info) TotalLength() int {
	return i.totalLength
}

// PieceLength returns the size of each piece
func (i Info) PieceLength() int {
	return i.pieceLength
}

// Files is a list of files that this torrent holds. Should not be empty
func (i Info) Files() []FileInfo {
	return i.files
}

func (i Info) File(idx int) *FileInfo {
	return i.filesByID[idx]
}

// SupportedFiles returns from all the files, the ones that have an extension supported by ffmpeg
func (i Info) SupportedFiles() []FileInfo {
	fi := make([]FileInfo, 0)
	for _, f := range i.files {
		_f := f
		if f.IsSupportedExtension() {
			fi = append(fi, _f)
		}
	}
	return fi
}

// FileInfo describes each file on the torrent. Is the second worst name in the project.
// Each file is identified by its position (which is important), the length and an arbitrary name
type FileInfo struct {
	idx    int
	length int
	name   string
	images []Image
}

// NewFileInfo creates a FileInfo
func NewFileInfo(idx int, length int, name string) (FileInfo, error) {
	return FileInfo{idx: idx, length: length, name: name, images: make([]Image, 0)}, nil
}

func (fi FileInfo) IsEqual(fi2 FileInfo) bool {
	return fi.idx == fi2.idx
}

// ID returns the ID, which is the index on the list of files of the torrent. zero indexed.
func (fi FileInfo) ID() int {
	return fi.idx
}

// Length returns the length of the file
func (fi FileInfo) Length() int {
	return fi.length
}

// Name returns the name of the file
func (fi FileInfo) Name() string {
	return fi.name
}

// DownloadSize is how much are we going to download from the file.
// Either a fixed amount or the whole file is smaller
func (fi FileInfo) DownloadSize() int {
	size := DownloadSize
	if size > fi.length {
		return fi.length
	}
	return DownloadSize
}

// IsSupportedExtension returns is the file has a supported extension to generate a preview
func (fi FileInfo) IsSupportedExtension() bool {
	supported := map[string]interface{}{
		".mp4": struct{}{},
	}

	ext := filepath.Ext(fi.name)
	_, found := supported[ext]
	return found
}

func (fi *FileInfo) AddImage(image Image) error {
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

func (fi FileInfo) Images() []Image {
	return fi.images
}
