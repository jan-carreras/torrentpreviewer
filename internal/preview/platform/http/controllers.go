package http

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/getTorrent"
	"prevtorrent/kit/command"
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

	c.IndentedJSON(200, getTorrentResponse{
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
		c.JSON(404, httpError{
			Message: err.Error(),
		})
		return
	}

	c.JSON(500, httpError{
		Message: fmt.Sprintf("Unexpected error: %v", err.Error()),
	})
}
