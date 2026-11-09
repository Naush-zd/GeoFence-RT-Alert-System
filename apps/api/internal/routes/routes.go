package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/geofencing-platform/apps/api/internal/handlers"
)

func New(api handlers.Handler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/healthz", api.Health)
	router.POST("/geofences", api.CreateGeofence)
	router.GET("/geofences", api.ListGeofences)
	router.POST("/vehicles", api.CreateVehicle)
	router.GET("/vehicles", api.ListVehicles)

	return router
}
