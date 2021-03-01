package preview_test

import (
	"context"
	"io/ioutil"
	"prevtorrent/internal/preview"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestMediaPart(t *testing.T) {
	torrentID := "ZOCmzqipffw7ollmic5hub6bpcsdeoqu"

	fi, err := preview.NewFileInfo(0, 1000, "movie.mp4")
	assert.NoError(t, err)
	files := []preview.File{fi}

	torrent, err := preview.NewInfo(torrentID, "test movie", 100, files, []byte("12345"))
	assert.NoError(t, err)

	pr := preview.NewPieceRange(torrent, fi, 0, 150, 100)
	data := []byte("1234")

	media := preview.NewMediaPart(torrentID, pr, data)

	assert.Equal(t, data, media.Data())
	assert.Equal(t, pr, media.PieceRange())
}

func TestBundlePlan_Bundle(t *testing.T) {
	type args struct {
		start  int
		offset int
		length int
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "I want all",
			args: args{
				start:  0,
				offset: 0,
				length: 100,
			},
			want: []byte("0123456789012345678912345111111111111111111111111122222222222222222222222229876543210987654321098765"),
		},
		{
			name: "Just the first part",
			args: args{
				start:  0,
				offset: 0,
				length: 10,
			},
			want: []byte("0123456789"),
		},
		{
			name: "The first part and half the second",
			args: args{
				start:  0,
				offset: 0,
				length: 15,
			},
			want: []byte("012345678901234"),
		},
		{
			name: "Second half of the file",
			args: args{
				start:  0,
				offset: 50,
				length: 50,
			},
			want: []byte("22222222222222222222222229876543210987654321098765"),
		},
		{
			name: "First Part and a half after half the file (using start instead of offset pretending is another file)",
			args: args{
				start:  50,
				offset: 0,
				length: 15,
			},
			want: []byte("222222222222222"),
		},
		{
			name: "Pretending that we have two files, second file with a little bit of offset",
			args: args{
				start:  70,
				offset: 10,
				length: 20,
			},
			want: []byte("43210987654321098765"),
		},
	}

	torrentID := "ZOCmzqipffw7ollmic5hub6bpcsdeoqu"

	fi, err := preview.NewFileInfo(0, 100, "movie.mp4")
	assert.NoError(t, err)
	files := []preview.File{fi}

	torrent, err := preview.NewInfo(torrentID, "test movie", 25, files, []byte(""))
	assert.NoError(t, err)

	ti := preview.NewTorrentImages(nil)

	plan := preview.NewDownloadPlan(torrent)
	assert.NoError(t, plan.AddAll(ti))

	part0 := []byte("0123456789012345678912345")
	part1 := []byte("1111111111111111111111111")
	part2 := []byte("2222222222222222222222222")
	part3 := []byte("9876543210987654321098765")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			registry, err := preview.NewPieceRegistry(context.Background(), fakeLogger(), plan, preview.NewPieceInMemoryStorage(*plan))
			assert.NoError(t, err)
			registry.RegisterPiece(preview.NewPiece(torrentID, 0, part0))
			registry.RegisterPiece(preview.NewPiece(torrentID, 1, part1))
			registry.RegisterPiece(preview.NewPiece(torrentID, 2, part2))
			registry.RegisterPiece(preview.NewPiece(torrentID, 3, part3))

			time.Sleep(time.Millisecond * 100) // Wait for the goroutines to spawn and queue the pieces

			pieceRange := preview.NewPieceRange(torrent, fi, tt.args.start, tt.args.offset, tt.args.length)

			bundlePlan := preview.NewBundlePlan()

			got, err := bundlePlan.Bundle(registry, pieceRange)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bundle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, string(tt.want), string(got.Data()))
		})
	}
}

func Test_TorrentImages(t *testing.T) {
	imgs := []preview.Image{
		preview.NewImage("torrentID", 0, "img1", 10),
		preview.NewImage("torrentID", 0, "img2", 12),
		preview.NewImage("torrentID", 1, "img1", 12),
	}
	images := preview.NewTorrentImages(imgs)
	assert.Equal(t, imgs, images.Images())
	assert.True(t, images.HaveImage("img1"))
	assert.False(t, images.HaveImage("img99"))
}

func fakeLogger() *logrus.Logger {
	l := logrus.New()
	l.Out = ioutil.Discard
	return l
}
