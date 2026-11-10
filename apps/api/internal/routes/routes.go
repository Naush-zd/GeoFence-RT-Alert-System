package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/yourusername/geofencing-platform/apps/api/internal/handlers"
)

func New(api handlers.Handler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	router.Use(cors.New(corsConfig))

	router.GET("/healthz", api.Health)
	router.POST("/geofences", api.CreateGeofence)
	router.GET("/geofences", api.ListGeofences)
	router.DELETE("/geofences/:geofence_id", api.DeleteGeofence)
	router.POST("/vehicles", api.CreateVehicle)
	router.GET("/vehicles", api.ListVehicles)
	router.POST("/vehicles/location", api.UpdateVehicleLocation)
	router.GET("/vehicles/location/:vehicle_id", api.GetVehicleLocation)
	router.POST("/alerts/configure", api.ConfigureAlert)
	router.GET("/alerts", api.ListAlerts)
	router.GET("/violations/history", api.ListViolations)
	router.GET("/ws/alerts", api.AlertsWebSocket)

	return router
}
