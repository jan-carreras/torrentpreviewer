package http

import (
	"prevtorrent/internal/platform/services"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run(s services.Services) error {
	server := NewServer(s)
	router := setupServer(server)
	return router.Run()
}

func setupServer(server *Server) *gin.Engine {
	router := gin.Default()
	router.Use(getCORS())
	router.GET("/torrent/:id", server.getTorrentController)
	router.POST("/unmagnetize", server.unmagnetizeController)
	router.POST("/torrent", server.newTorrentController)
	return router
}

func getCORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:  []string{"*", "localhost"},
		AllowMethods:  []string{"GET", "POST"},
		AllowHeaders:  []string{"Origin", "Cache-Control", "X-Requested-With"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	})
}
