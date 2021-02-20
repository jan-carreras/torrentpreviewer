package http

import (
	"errors"
	"fmt"
	"prevtorrent/internal/preview"
	"prevtorrent/kit/command"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

/**
TODO: Well. This is trash. It's a quick and dirty HTTP server that serves the information of a torrent
      so that I can start working on a frontend solution. Am I ashamed about it? A little bit. Does it
	  has tests? Of course not. Do we need it for the MVP? Yup.
TODO: The backend is SQLite. So... expect 0 performance and don't you dare, Jan from the future, to
      complain about it.
*/

func Run(c Container, bus command.Bus) error {
	server := NewServer(c, bus)
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET"},
		AllowHeaders:  []string{"Origin"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))
	router.GET("/torrent/:id", server.getTorrent)

	// TODO: Receive this from configuration
	router.Use(static.Serve("/image", static.LocalFile("./tmp/images", false)))

	return router.Run()
}

type Server struct {
	container Container
	bus       command.Bus
}

func NewServer(container Container, bus command.Bus) *Server {
	return &Server{container: container, bus: bus}
}

func (s *Server) getTorrent(c *gin.Context) {
	torrent, err := s.container.TorrentRepo.Get(c, c.Params.ByName("id"))
	if err != nil {
		if errors.Is(err, preview.ErrNotFound) {
			c.JSON(404, httpError{
				Message: err.Error(),
			})
			return
		}

		c.JSON(500, httpError{
			Message: fmt.Sprintf("Unexpected error: %v", err.Error()),
		})
		return
	}

	images, err := s.container.ImageRepository.ByTorrent(c, torrent.ID())
	if err != nil {
		c.JSON(500, httpError{
			Message: fmt.Sprintf("Unexpected error: %v", err.Error()),
		})
		return
	}

	cache := make(map[int]Image) // TODO: This only supports one image per File
	for _, img := range images.Images() {
		if img.Length() == 0 {
			continue
		}
		cache[img.FileID()] = Image{Src: img.Name()}
	}

	files := make([]File, 0, len(torrent.Files()))
	for _, f := range torrent.Files() {
		images := make([]Image, 0)
		if image, found := cache[f.ID()]; found {
			images = append(images, image)
		}

		files = append(files, File{
			ID:     f.ID(),
			Length: f.Length(),
			Name:   f.Name(),
			Images: images,
		})
	}

	c.IndentedJSON(200, getTorrentResponse{
		Torrent: Torrent{
			Id:     torrent.ID(),
			Name:   torrent.Name(),
			Length: torrent.TotalLength(),
			Files:  files,
		},
	})
}
