package downloadPartials_test

import (
	"context"
	"errors"
	"io/ioutil"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/platform/client/clientmocks"
	"prevtorrent/internal/preview/platform/storage/storagemocks"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_DownloadPartials_GetTorrentError(t *testing.T) {
	torrentID := "zocmzqipffw7ollmic5hub6bpcsdeoqu"

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, torrentID).Return(preview.Info{}, errors.New("fake error"))

	torrentDownloader := new(clientmocks.TorrentDownloader)

	imageExtractor := new(storagemocks.ImageExtractor)

	imagePersister := new(storagemocks.ImagePersister)

	imageRepository := new(storagemocks.ImageRepository)

	service := downloadPartials.NewService(
		fakeLogger(),
		torrentRepository,
		torrentDownloader,
		imageExtractor,
		imagePersister,
		imageRepository,
	)

	cmd := downloadPartials.CMD{ID: torrentID}
	err := service.DownloadPartials(context.Background(), cmd)
	assert.Error(t, err)
}

func TestService_DownloadPartials_DownloadPartsFails(t *testing.T) {
	torrentID := "zocmzqipffw7ollmic5hub6bpcsdeoqu"

	f, err := preview.NewFileInfo(0, 100, "video.mp4")
	assert.NoError(t, err)

	var files []preview.FileInfo
	files = append(files, f)

	torrent, err := preview.NewInfo(torrentID, "test torrent", 100, files, []byte("torrent-data"))
	assert.NoError(t, err)

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, torrentID).Return(torrent, nil)

	torrentDownloader := new(clientmocks.TorrentDownloader)
	torrentDownloader.On("DownloadParts", mock.Anything, mock.Anything).
		Return(nil, errors.New("error when downloading"))

	imageExtractor := new(storagemocks.ImageExtractor)

	imagePersister := new(storagemocks.ImagePersister)

	imageRepository := new(storagemocks.ImageRepository)
	imageRepository.On("ByTorrent", mock.Anything, torrentID).
		Return(preview.NewTorrentImages(nil), nil)

	service := downloadPartials.NewService(
		fakeLogger(),
		torrentRepository,
		torrentDownloader,
		imageExtractor,
		imagePersister,
		imageRepository,
	)

	cmd := downloadPartials.CMD{ID: torrentID}
	err = service.DownloadPartials(context.Background(), cmd)
	assert.Error(t, err)
}

func TestService_DownloadPartials_RegistryClosesWithNoParts(t *testing.T) {
	torrentID := "zocmzqipffw7ollmic5hub6bpcsdeoqu"

	f, err := preview.NewFileInfo(0, 100, "video.mp4")
	assert.NoError(t, err)

	var files []preview.FileInfo
	files = append(files, f)

	torrent, err := preview.NewInfo(torrentID, "test torrent", 100, files, []byte("torrent-data"))
	assert.NoError(t, err)

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, torrentID).Return(torrent, nil)

	torrentImages := preview.NewTorrentImages(nil)

	plan := preview.NewDownloadPlan(torrent, torrentImages)
	err = plan.AddAll()
	assert.NoError(t, err)

	registry, err := preview.NewPieceRegistry(context.Background(), plan, preview.NewPieceInMemoryStorage(*plan))
	assert.NoError(t, err)
	registry.NoMorePieces()

	torrentDownloader := new(clientmocks.TorrentDownloader)
	torrentDownloader.On("DownloadParts", mock.Anything, mock.Anything).
		Return(registry, nil)

	imageExtractor := new(storagemocks.ImageExtractor)

	imageRepository := new(storagemocks.ImageRepository)
	imageRepository.On("ByTorrent", mock.Anything, torrentID).
		Return(preview.NewTorrentImages(nil), nil)

	imagePersister := new(storagemocks.ImagePersister)

	service := downloadPartials.NewService(
		fakeLogger(),
		torrentRepository,
		torrentDownloader,
		imageExtractor,
		imagePersister,
		imageRepository,
	)

	cmd := downloadPartials.CMD{ID: torrentID}
	err = service.DownloadPartials(context.Background(), cmd)
	assert.NoError(t, err)
}

func TestService_DownloadPartials_ExtractImageFails(t *testing.T) {
	torrentID := "zocmzqipffw7ollmic5hub6bpcsdeoqu"

	f, err := preview.NewFileInfo(0, 10, "video.mp4")
	assert.NoError(t, err)

	var files []preview.FileInfo
	files = append(files, f)

	torrent, err := preview.NewInfo(torrentID, "test torrent", 5, files, []byte("torrent-data"))
	assert.NoError(t, err)

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, torrentID).Return(torrent, nil)

	torrentImages := preview.NewTorrentImages(nil)

	plan := preview.NewDownloadPlan(torrent, torrentImages)
	assert.NoError(t, plan.AddAll())
	registry, err := preview.NewPieceRegistry(context.Background(), plan, preview.NewPieceInMemoryStorage(*plan))
	assert.NoError(t, err)
	registry.RegisterPiece(preview.NewPiece(torrentID, 0, []byte("12345")))
	registry.RegisterPiece(preview.NewPiece(torrentID, 1, []byte("67890")))
	registry.NoMorePieces()

	time.Sleep(time.Millisecond * 100) // Give some time for the events to be process in the goroutine

	torrentDownloader := new(clientmocks.TorrentDownloader)
	torrentDownloader.On("DownloadParts", mock.Anything, mock.Anything).
		Return(registry, nil)

	imageExtractor := new(storagemocks.ImageExtractor)
	imageExtractor.On("ExtractImage", mock.Anything, []byte("1234567890"), 5).Return(nil, errors.New("fake image error"))

	imageRepository := new(storagemocks.ImageRepository)
	imageRepository.On("ByTorrent", mock.Anything, torrentID).
		Return(preview.NewTorrentImages(nil), nil)

	imagePersister := new(storagemocks.ImagePersister)

	service := downloadPartials.NewService(
		fakeLogger(),
		torrentRepository,
		torrentDownloader,
		imageExtractor,
		imagePersister,
		imageRepository,
	)

	cmd := downloadPartials.CMD{ID: torrentID}
	err = service.DownloadPartials(context.Background(), cmd)
	assert.Error(t, err)
}

func TestService_DownloadPartials_PersistingImageFails(t *testing.T) {
	torrentID := "zocmzqipffw7ollmic5hub6bpcsdeoqu"

	f, err := preview.NewFileInfo(0, 10, "video.mp4")
	assert.NoError(t, err)

	var files []preview.FileInfo
	files = append(files, f)

	torrent, err := preview.NewInfo(torrentID, "test torrent", 5, files, []byte("torrent-data"))
	assert.NoError(t, err)

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, torrentID).Return(torrent, nil)

	torrentImages := preview.NewTorrentImages(nil)
	plan := preview.NewDownloadPlan(torrent, torrentImages)
	assert.NoError(t, plan.AddAll())
	registry, err := preview.NewPieceRegistry(context.Background(), plan, preview.NewPieceInMemoryStorage(*plan))
	assert.NoError(t, err)
	registry.RegisterPiece(preview.NewPiece(torrentID, 0, []byte("12345")))
	registry.RegisterPiece(preview.NewPiece(torrentID, 1, []byte("67890")))
	registry.NoMorePieces()

	time.Sleep(time.Millisecond * 100) // Give some time for the events to be process in the goroutine

	torrentDownloader := new(clientmocks.TorrentDownloader)
	torrentDownloader.On("DownloadParts", mock.Anything, mock.Anything).
		Return(registry, nil)

	imgBytes := []byte("JPG binary data here")
	imageExtractor := new(storagemocks.ImageExtractor)
	imageExtractor.On("ExtractImage", mock.Anything, []byte("1234567890"), 5).Return(imgBytes, nil)

	imageRepository := new(storagemocks.ImageRepository)
	imageRepository.On("ByTorrent", mock.Anything, torrentID).
		Return(preview.NewTorrentImages(nil), nil)

	imagePersister := new(storagemocks.ImagePersister)
	imagePersister.On("PersistFile", mock.Anything, "zocmzqipffw7ollmic5hub6bpcsdeoqu.0.0-1.video.mp4.jpg", imgBytes).
		Return(errors.New("fake storing error"))

	service := downloadPartials.NewService(
		fakeLogger(),
		torrentRepository,
		torrentDownloader,
		imageExtractor,
		imagePersister,
		imageRepository,
	)

	cmd := downloadPartials.CMD{ID: torrentID}
	err = service.DownloadPartials(context.Background(), cmd)
	assert.Error(t, err)
}

func TestService_DownloadPartials_BaseCase(t *testing.T) {
	torrentID := "zocmzqipffw7ollmic5hub6bpcsdeoqu"

	f, err := preview.NewFileInfo(0, 10, "video.mp4")
	assert.NoError(t, err)

	var files []preview.FileInfo
	files = append(files, f)

	torrent, err := preview.NewInfo(torrentID, "test torrent", 5, files, []byte("torrent-data"))
	assert.NoError(t, err)

	torrentRepository := new(storagemocks.TorrentRepository)
	torrentRepository.On("Get", mock.Anything, torrentID).Return(torrent, nil)

	torrentImages := preview.NewTorrentImages(nil)
	plan := preview.NewDownloadPlan(torrent, torrentImages)
	assert.NoError(t, plan.AddAll())
	registry, err := preview.NewPieceRegistry(context.Background(), plan, preview.NewPieceInMemoryStorage(*plan))
	assert.NoError(t, err)
	registry.RegisterPiece(preview.NewPiece(torrentID, 0, []byte("12345")))
	registry.RegisterPiece(preview.NewPiece(torrentID, 1, []byte("67890")))
	registry.NoMorePieces()

	time.Sleep(time.Millisecond * 100) // Give some time for the events to be process in the goroutine

	torrentDownloader := new(clientmocks.TorrentDownloader)
	torrentDownloader.On("DownloadParts", mock.Anything, mock.Anything).
		Return(registry, nil)

	imgBytes := []byte("JPG binary data here")
	imageExtractor := new(storagemocks.ImageExtractor)
	imageExtractor.On("ExtractImage", mock.Anything, []byte("1234567890"), 5).Return(imgBytes, nil)

	img := preview.NewImage(
		torrentID,
		0,
		"zocmzqipffw7ollmic5hub6bpcsdeoqu.0.0-1.video.mp4.jpg",
		len(imgBytes),
	)

	imageRepository := new(storagemocks.ImageRepository)
	imageRepository.On("ByTorrent", mock.Anything, torrentID).
		Return(preview.NewTorrentImages(nil), nil)
	imageRepository.On("Persist", mock.Anything, img).
		Return(nil)

	imagePersister := new(storagemocks.ImagePersister)
	imagePersister.On("PersistFile", mock.Anything, "zocmzqipffw7ollmic5hub6bpcsdeoqu.0.0-1.video.mp4.jpg", imgBytes).
		Return(nil)

	service := downloadPartials.NewService(
		fakeLogger(),
		torrentRepository,
		torrentDownloader,
		imageExtractor,
		imagePersister,
		imageRepository,
	)

	cmd := downloadPartials.CMD{ID: torrentID}
	err = service.DownloadPartials(context.Background(), cmd)
	assert.NoError(t, err)
}

func fakeLogger() *logrus.Logger {
	l := logrus.New()
	l.Out = ioutil.Discard
	return l
}
