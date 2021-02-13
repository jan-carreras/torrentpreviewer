package preview

import (
	"fmt"
	"strings"
)

type DownloadPlan struct {
	torrent     Info
	pieceRanges []PieceRange
}

func (dp *DownloadPlan) GetTorrent() Info {
	return dp.torrent
}

func NewDownloadPlan(torrent Info) *DownloadPlan {
	return &DownloadPlan{
		torrent:     torrent,
		pieceRanges: make([]PieceRange, 0),
	}
}
func (dp *DownloadPlan) GetPlan() []PieceRange {
	return dp.pieceRanges
}

func (dp *DownloadPlan) CountPars() int {
	count := 0
	for _, p := range dp.pieceRanges {
		count += p.end - p.start
	}
	return count
}

func (dp *DownloadPlan) Download(fi FileInfo, length, offset int) error {
	if length+offset > fi.length {
		return fmt.Errorf("length+offset should be less than the total size of the file. having=%v, expecting <= %v", length+offset, fi.length)
	}
	if length <= 0 {
		return fmt.Errorf("only valid positive non-zero lengths. having=%v", length)
	}
	if length < 0 {
		return fmt.Errorf("only valid positive or zero offsets. having=%v", offset)
	}

	if !fi.IsSupportedExtension() {
		return fmt.Errorf("file %s has not a supported extension", fi.name)
	}
	start := findStartingByteOfFile(dp.torrent, fi)

	dp.addToDownloadPlan(NewPieceRange(dp.torrent, fi, start, offset, length))
	return nil
}

func (dp *DownloadPlan) addToDownloadPlan(piece PieceRange) {
	dp.pieceRanges = append(dp.pieceRanges, piece)
}

type PieceRange struct {
	fi               FileInfo
	start            int // piece start. The file we want to download starts in this Piece
	end              int // piece end. The file we want to download ends in this Piece
	firstPieceOffset int // The file not necessarily starts at the byte 0 of the Piece. This offset indicates when it starts inside the piece
	lastPieceOffset  int // The file not necessarily ends at the end of the last Piece. This offset indicates when it ends inside the piece
	pieceLength      int // The length of each piece of this torrent
}

func NewPieceRange(t Info, fi FileInfo, start, offset, length int) PieceRange {
	return PieceRange{
		fi:               fi,
		start:            (start + offset) / t.pieceLength,
		end:              (start + offset + length) / t.pieceLength,
		firstPieceOffset: (start + offset) % t.pieceLength,
		lastPieceOffset:  (start + offset + length) % t.pieceLength,
		pieceLength:      t.pieceLength,
	}
}

func (p PieceRange) Name() string {
	return strings.ReplaceAll(p.fi.name, "/", "--")
}

func (p PieceRange) Start() int {
	return p.start
}

func (p PieceRange) End() int {
	return p.end
}

func (p PieceRange) StartOffset(idx int) int {
	if idx == p.start {
		return p.firstPieceOffset
	}
	return 0
}

func (p PieceRange) EndOffset(idx int) int {
	if idx == p.end {
		return p.lastPieceOffset
	}
	return p.pieceLength
}

func findStartingByteOfFile(t Info, fi FileInfo) int {
	start := 0
	for _, f := range t.files {
		if f == fi {
			break
		}
		start += f.length
	}
	return start
}

type DownloadedPart struct {
	torrentID  string
	pieceRange PieceRange
	data       []byte
}

func NewDownloadedPart(torrentID string, pieceRange PieceRange, data []byte) DownloadedPart {
	return DownloadedPart{torrentID: torrentID, pieceRange: pieceRange, data: data}
}

func (p DownloadedPart) Name() string {
	return fmt.Sprintf("%v.%v.%v-%v.%v.jpg",
		p.torrentID,
		p.pieceRange.fi.idx,
		p.pieceRange.Start(),
		p.pieceRange.End(),
		p.pieceRange.Name(),
	)
}

func (p DownloadedPart) PieceRange() PieceRange {
	return p.pieceRange
}

func (p DownloadedPart) Data() []byte {
	return p.data
}
