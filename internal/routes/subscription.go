package routes

import (
	"github.com/gin-gonic/gin"
	"calenduh-backend/internal/handlers"
)

func RegisterSubscriptionRoutes(router *gin.Engine) {
	subRoutes := router.Group("/sub")
	{
		subRoutes.POST("/subscribe", handlers.Subscribe)
		subRoutes.POST("/unsubscribe", handlers.Unsubscribe)
		subRoutes.GET("/fetchAllSubbedCalendars", handlers.FetchAllSubbedCalendars)
		// POST, GET, DELETE, PUT-all fields update, PATCH-certain selected fields update
	}
}