package preview

import (
	"bytes"
	"errors"
	"fmt"
)

// MediaPart has the binary data expected after downloading a PieceRange from the
// DownloadPlan. So, usually it's going to be a partial video from a Torrent.
type MediaPart struct {
	torrentID  string
	pieceRange PieceRange
	data       []byte
}

// NewMediaPart creates a MediaPart
func NewMediaPart(torrentID string, pieceRange PieceRange, data []byte) MediaPart {
	return MediaPart{torrentID: torrentID, pieceRange: pieceRange, data: data}
}

// PieceRange returns the obvious
func (p MediaPart) PieceRange() PieceRange {
	return p.pieceRange
}

// Data raw data of the file
func (p MediaPart) Data() []byte {
	return p.data
}

type pieceRangeCounter struct {
	pieceRange       PieceRange
	piecesDownloaded int
}

func newPieceRangeCounter(
	pieceRange PieceRange,
) *pieceRangeCounter {
	return &pieceRangeCounter{pieceRange: pieceRange}
}

func (c *pieceRangeCounter) addOne() {
	c.piecesDownloaded++
}

func (c *pieceRangeCounter) areAllPiecesDownloaded() bool {
	return c.piecesDownloaded >= c.pieceRange.PieceCount()
}

// BundlePlan gets a PieceRange which is the definition of a file we want to download,
// and a PieceRegistry which is were we have stored individual pieces,
// and reads the whole file from the pieces.
// We have to remember that a piece might contain multiple files. Once piece is not a file
// and each files does not start at the start of a piece. That would be coincidence.
// BundlePlan takes care of this logic and returns a MediaPart, which is the closes thing
// of a file we're going to have.
type BundlePlan struct{}

// NewBundlePlan creates a BundlePlan
func NewBundlePlan() BundlePlan {
	return BundlePlan{}
}

// Bundle transform a PieceRange (the file we want) to a MediaPart (the actual file)
func (b BundlePlan) Bundle(registry *PieceRegistry, pieceRange PieceRange) (MediaPart, error) {
	piece := new(bytes.Buffer)

	for pieceIdx := pieceRange.Start(); pieceIdx <= pieceRange.End(); pieceIdx++ {
		p, found := registry.GetPiece(pieceIdx)
		if !found {
			return MediaPart{}, errors.New("piece not found in the registry. could be ignored but kept to further investigate")
		}
		start := pieceRange.StartOffset(pieceIdx)
		end := pieceRange.EndOffset(pieceIdx)
		if start > end {
			return MediaPart{}, fmt.Errorf("start offset %v bigger than end offset %v", start, end)
		}
		if start > len(p.data) {
			return MediaPart{}, fmt.Errorf("start offset %v is bigger than length of slice %v", start, len(p.data))
		}
		if end > len(p.data) {
			return MediaPart{}, fmt.Errorf("end offset %v is bigger than length of slice %v", start, len(p.data))
		}

		rawData := p.data[start:end]
		_, err := piece.Write(rawData)
		if err != nil {
			return MediaPart{}, err
		}
	}

	return NewMediaPart(pieceRange.Torrent().ID(), pieceRange, piece.Bytes()), nil
}

// TorrentImages represents all the images of a torrent
type TorrentImages struct {
	images    []Image
	imageName map[string]interface{}
}

// NewTorrentImages returns a TorrentImages
func NewTorrentImages(images []Image) *TorrentImages {
	imageName := make(map[string]interface{})
	for _, img := range images {
		imageName[img.name] = struct{}{}
	}
	return &TorrentImages{images: images, imageName: imageName}
}

func (a *TorrentImages) Images() []Image {
	return a.images
}

// HaveImage stupid name to check if we already have the filename
func (a *TorrentImages) HaveImage(name string) bool {
	_, found := a.imageName[name]
	return found
}
