package routes

import (
	"github.com/gin-gonic/gin"
	"calenduh-backend/internal/handlers"
)

func RegisterGroupRoutes(router *gin.Engine) {
	subRoutes := router.Group("/sub")
	{
		subRoutes.POST("/createGroup", handlers.CreateGroup)
		subRoutes.GET("/fetchGroup", handlers.FetchGroup)
		subRoutes.GET("/fetchAllGroups", handlers.FetchAllGroups)
		subRoutes.PATCH("/updateGroup", handlers.UpdateGroup)
		// POST, GET, DELETE, PUT-all fields update, PATCH-certain selected fields update
	}
}