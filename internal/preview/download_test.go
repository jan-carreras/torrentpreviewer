package preview_test

import (
	"prevtorrent/internal/preview"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Piece(t *testing.T) {
	torrentdID := "zocmzqipffw7ollmic5hub6bpcsdeoqu"
	pieceID := 55
	data := []byte("1234")
	p := preview.NewPiece(torrentdID, pieceID, data)

	assert.Equal(t, torrentdID, p.TorrentID())
	assert.Equal(t, pieceID, p.ID())
	assert.Equal(t, data, p.Data())
}

func TestPieceRange(t *testing.T) {
	torrentID := "zocmzqipffw7ollmic5hub6bpcsdeoqu"
	pieceLength := 100

	fi, err := preview.NewFileInfo(0, 1000, "test/movie.mp4")
	assert.NoError(t, err)
	fi2, err := preview.NewFileInfo(0, 150, "test/movie2.mp4")
	assert.NoError(t, err)
	torrent, err := preview.NewInfo(torrentID, "generic movie", pieceLength, []preview.FileInfo{fi, fi2}, []byte(""))
	assert.NoError(t, err)

	type args struct {
		t      preview.Info
		fi     preview.FileInfo
		start  int
		offset int
		length int
	}
	type want struct {
		name             string
		pieceStart       int
		pieceEnd         int
		startOffsetBytes int
		endOffsetBytes   int
		pieceCount       int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "first file, from the start 50 bytes",
			args: args{
				t:      torrent,
				fi:     fi,
				start:  0,
				offset: 0,
				length: 50,
			},
			want: want{
				name:             "zocmzqipffw7ollmic5hub6bpcsdeoqu.0.0-0.test--movie.mp4.jpg",
				pieceStart:       0,
				pieceEnd:         0,
				startOffsetBytes: 0,
				endOffsetBytes:   50,
				pieceCount:       1,
			},
		},
		{
			name: "first file read 100 bytes",
			args: args{
				t:      torrent,
				fi:     fi,
				start:  0,
				offset: 0,
				length: 100,
			},
			want: want{
				name:             "zocmzqipffw7ollmic5hub6bpcsdeoqu.0.0-0.test--movie.mp4.jpg",
				pieceStart:       0,
				pieceEnd:         0,
				startOffsetBytes: 0,
				endOffsetBytes:   100,
				pieceCount:       1,
			},
		},
		{
			name: "first file, skip 25 bytes and read the rest of the block",
			args: args{
				t:      torrent,
				fi:     fi,
				start:  0,
				offset: 25,
				length: 75,
			},
			want: want{
				name:             "zocmzqipffw7ollmic5hub6bpcsdeoqu.0.0-0.test--movie.mp4.jpg",
				pieceStart:       0,
				pieceEnd:         0,
				startOffsetBytes: 25,
				endOffsetBytes:   100,
				pieceCount:       1,
			},
		},
		{
			name: "first file, skip 25 bytes and read the rest of the block and one entire more",
			args: args{
				t:      torrent,
				fi:     fi,
				start:  0,
				offset: 25,
				length: 175,
			},
			want: want{
				name:             "zocmzqipffw7ollmic5hub6bpcsdeoqu.0.0-1.test--movie.mp4.jpg",
				pieceStart:       0,
				pieceEnd:         1,
				startOffsetBytes: 25,
				endOffsetBytes:   100,
				pieceCount:       2,
			},
		},
		{
			name: "complicated one. Second file",
			args: args{
				t:      torrent,
				fi:     fi,
				start:  1000,
				offset: 50,
				length: 100,
			},
			want: want{
				name:             "zocmzqipffw7ollmic5hub6bpcsdeoqu.0.10-11.test--movie.mp4.jpg",
				pieceStart:       10,
				pieceEnd:         11,
				startOffsetBytes: 50,
				endOffsetBytes:   50,
				pieceCount:       2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := preview.NewPieceRange(tt.args.t, tt.args.fi, tt.args.start, tt.args.offset, tt.args.length)

			assert.Equal(t, torrent, got.Torrent())
			assert.Equal(t, tt.want.name, got.Name())
			assert.Equal(t, tt.want.pieceStart, got.Start())
			assert.Equal(t, tt.want.pieceEnd, got.End())
			assert.Equal(t, tt.want.startOffsetBytes, got.StartOffset(got.Start()))
			assert.Equal(t, tt.want.endOffsetBytes, got.EndOffset(got.End()))
			assert.Equal(t, tt.want.pieceCount, got.PieceCount())
		})
	}
}

func TestDownloadPlan_GetTorrent(t *testing.T) {
	torrentID := "zocmzqipffw7ollmic5hub6bpcsdeoqu"

	fi, err := preview.NewFileInfo(0, 1000, "movie.mp4")
	assert.NoError(t, err)
	torrent, err := preview.NewInfo(torrentID, "generic movie", 100, []preview.FileInfo{fi}, []byte(""))
	assert.NoError(t, err)

	torrentImages := preview.NewTorrentImages(nil)

	plan := preview.NewDownloadPlan(torrent, torrentImages)

	assert.Equal(t, torrent, plan.GetTorrent())
}

func TestDownloadPlan_GetPlan(t *testing.T) {
	torrentID := "zocmzqipffw7ollmic5hub6bpcsdeoqu"

	fi, err := preview.NewFileInfo(0, 1000, "movie.mp4")
	assert.NoError(t, err)
	torrent, err := preview.NewInfo(torrentID, "generic movie", 100, []preview.FileInfo{fi}, []byte(""))
	assert.NoError(t, err)

	torrentImages := preview.NewTorrentImages(nil)

	plan := preview.NewDownloadPlan(torrent, torrentImages)

	assert.Equal(t, make([]preview.PieceRange, 0), plan.GetPlan())
}

func TestDownloadPlan_AddAll(t *testing.T) {
	torrentID := "zocmzqipffw7ollmic5hub6bpcsdeoqu"

	fi, err := preview.NewFileInfo(0, 1000, "movie.mp4")
	assert.NoError(t, err)
	f2, err := preview.NewFileInfo(0, 500, "movie2.mp4")
	assert.NoError(t, err)
	torrent, err := preview.NewInfo(torrentID, "generic movie", 100, []preview.FileInfo{fi, f2}, []byte(""))
	assert.NoError(t, err)

	torrentImages := preview.NewTorrentImages(nil)

	plan := preview.NewDownloadPlan(torrent, torrentImages)
	err = plan.AddAll()
	assert.NoError(t, err)

	assert.Equal(t, 15, plan.CountPieces())

	pieceRanges := plan.GetPlan()

	assert.Len(t, pieceRanges, 2)
}
