package http

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"prevtorrent/internal/platform/services"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run(s services.Services, bus *cqrs.CommandBus, eventBus *cqrs.EventBus) error {
	server := NewServer(s, bus, eventBus)
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*", "localhost"},
		AllowMethods:  []string{"GET", "POST"},
		AllowHeaders:  []string{"Origin", "Cache-Control", "X-Requested-With"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))
	router.GET("/torrent/:id", server.getTorrentController)
	router.POST("/unmagnetize", server.unmagnetizeController)
	router.POST("/torrent", server.newTorrentController)

	return router.Run()
}
