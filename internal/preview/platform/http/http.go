package http

import (
	"prevtorrent/kit/command"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func Run(c Container, bus command.Bus) error {
	server := NewServer(c.services, bus)
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET"},
		AllowHeaders:  []string{"Origin"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))
	router.GET("/torrent/:id", server.getTorrentController)
	router.POST("/unmagnetize", server.unmagnetizeController)

	// TODO: Move this logic to nginx
	router.Use(static.Serve("/image", static.LocalFile(c.Config.ImageDir, false)))

	return router.Run()
}
