package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/getTorrent"
	"prevtorrent/internal/preview/unmagnetize"
	"prevtorrent/kit/command"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	services Services
	bus      command.Bus
}

func NewServer(services Services, bus command.Bus) *Server {
	return &Server{services: services, bus: bus}
}

func (s *Server) getTorrentController(c *gin.Context) {
	torrent, err := s.services.GetTorrent.Get(c, getTorrent.CMD{
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

func (s *Server) unmagnetizeController(c *gin.Context) {
	magnet := c.PostForm("magnet")
	if len(magnet) == 0 {
		c.JSON(http.StatusBadRequest, httpError{
			Message: fmt.Sprintf("magnet link cannot be empty"),
		})
		return
	}
	ctxWithCancellation, cancel := context.WithTimeout(c, time.Second*15)
	defer cancel()
	torrentID, err := s.services.Unmagnetize.Handle(ctxWithCancellation, unmagnetize.CMD{Magnet: magnet})
	if err != nil {
		s.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": torrentID,
	})
}
