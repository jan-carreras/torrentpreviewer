package preview

import (
	"fmt"
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

	dp.pieceRanges = append(dp.pieceRanges, PieceRange{
		fi:               fi,
		start:            (start + offset) / dp.torrent.pieceLength,
		end:              (start + offset + length) / dp.torrent.pieceLength,
		firstPieceOffset: (start + offset) % dp.torrent.pieceLength,
		lastPieceOffset:  (start + offset + length) % dp.torrent.pieceLength,
		pieceLength:      dp.torrent.pieceLength,
	})
	return nil
}

// TODO: Refactor this object pretty please
type PieceRange struct {
	fi               FileInfo
	start            int // piece start
	end              int // piece end
	firstPieceOffset int
	lastPieceOffset  int
	pieceLength      int
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
	pieceRange PieceRange
	data       []byte
}

func NewDownloadedPart(pieceRange PieceRange, data []byte) DownloadedPart {
	return DownloadedPart{pieceRange: pieceRange, data: data}
}

func (p DownloadedPart) PieceRange() PieceRange {
	return p.pieceRange
}

func (p DownloadedPart) Data() []byte {
	return p.data
}
