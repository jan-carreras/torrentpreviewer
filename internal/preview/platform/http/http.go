package http

import (
	http2 "net/http"
	"prevtorrent/internal/platform/services"
	"time"

	"github.com/cnjack/throttle"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run(s services.Services) *http2.Server {
	server := NewServer(s)
	router := setupServer(server)

	srv := &http2.Server{
		Addr:    ":8080",
		Handler: router,
	}

	return srv
}

func setupServer(server *Server) *gin.Engine {
	router := gin.Default()
	router.Use(getCORS())

	router.Use(throttle.Policy(&throttle.Quota{Limit: 4, Within: time.Second}))
	router.Use(throttle.Policy(&throttle.Quota{Limit: 120, Within: time.Minute}))

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
