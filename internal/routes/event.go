package routes

import (
	"github.com/gin-gonic/gin"
	"calenduh-backend/internal/handlers"
)

func RegisterEventRoutes(router *gin.Engine) {
	eventRoutes := router.Group("/event")
	{
		eventRoutes.POST("/createEvent", handlers.CreateEvent)
		eventRoutes.GET("/fetchEvent", handlers.FetchEvent)
		eventRoutes.GET("/fetchEventName", handlers.FetchEventName)
		eventRoutes.GET("/fetchAllEvents", handlers.FetchAllEvents)
		eventRoutes.PATCH("/updateEvent", handlers.UpdateEvent)
		// POST, GET, DELETE, PUT-all fields update, PATCH-certain selected fields update
	}
}
