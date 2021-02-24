package http

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"prevtorrent/internal/platform/services"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/getTorrent"
	"prevtorrent/internal/preview/importTorrent"
	"prevtorrent/internal/preview/unmagnetize"
	"prevtorrent/kit/command"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	services services.Services
	bus      command.Bus
}

func NewServer(services services.Services, bus command.Bus) *Server {
	return &Server{services: services, bus: bus}
}

func (s *Server) getTorrentController(c *gin.Context) {
	torrent, err := s.services.GetTorrent().Get(c, getTorrent.CMD{
		TorrentID: c.Params.ByName("id"),
	})

	if err != nil {
		s.handleError(c, err)
		return
	}

	c.IndentedJSON(http.StatusOK, getTorrentResponse{
		Torrent: Torrent{
			Id:     torrent.ID(),
			Name:   torrent.Name(),
			Length: torrent.TotalLength(),
			Files:  makeFiles(torrent),
		},
	})
}

func makeFiles(torrent preview.Info) []File {
	files := make([]File, 0, len(torrent.Files()))
	for _, f := range torrent.Files() {
		images := make([]Image, 0)
		for _, img := range f.Images() {
			images = append(images, Image{
				Src:     img.Name(),
				Length:  img.Length(),
				IsValid: img.Length() != 0,
			})
		}

		files = append(files, File{
			ID:          f.ID(),
			Length:      f.Length(),
			Name:        f.Name(),
			Images:      images,
			IsSupported: f.IsSupportedExtension(),
		})
	}
	return files
}

func (s *Server) handleError(c *gin.Context, err error) {
	if errors.Is(err, preview.ErrNotFound) {
		c.JSON(http.StatusNotFound, httpError{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, httpError{
		Message: fmt.Sprintf("Unexpected error: %v", err.Error()),
	})
}

func (s *Server) unmagnetizeController(ctx *gin.Context) {
	magnet := ctx.PostForm("magnet")
	if len(magnet) == 0 {
		ctx.JSON(http.StatusBadRequest, httpError{
			Message: "magnet link cannot be empty",
		})
		return
	}
	ctxWithCancellation, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	torrentID, err := s.services.Unmagnetize().Handle(ctxWithCancellation, unmagnetize.CMD{Magnet: magnet})
	if err != nil {
		s.handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"id": torrentID,
	})
}

func (s *Server) newTorrentController(ctx *gin.Context) {
	readFile := func() ([]byte, error) {
		file, err := ctx.FormFile("file")
		if err != nil {
			return nil, err
		}
		f, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer f.Close()

		return ioutil.ReadAll(f)
	}

	file, err := readFile()
	if err != nil {
		s.handleError(ctx, err)
		return
	}

	torrent, err := s.services.ImportTorrent().Import(ctx, importTorrent.CMD{
		TorrentRaw: file,
	})
	if err != nil {
		s.handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id": torrent.ID(),
	})
}
