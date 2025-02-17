package main

import (
	"calenduh-backend/internal/controllers"
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/util"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	timeStarted := time.Now()

	client := database.New()

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	env := util.GetEnv("GO_ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Router setup
	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(controllers.Authorize)
	router.GET("/health", func(c *gin.Context) {
		uptime := time.Since(timeStarted).Truncate(time.Second)
		message := gin.H{
			"uptime":       fmt.Sprintf("%v", uptime),
			"active_users": util.ActiveUsers.ItemCount(),
			"daily_users":  util.DailyUsers.ItemCount(),
		}
		c.PureJSON(http.StatusOK, message)
		return
	})

	// Setup Routes
	setupRoutes(router)

	// Signal handling
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal("router failed to run:", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	cleanup(server)

	// Shutdown HTTP server gracefully
}

func setupRoutes(router *gin.Engine) {
	authentication := router.Group("/auth")
	users := router.Group("/users")
	events := router.Group("/event")
	groups := router.Group("/groups")
	calendars := router.Group("/calendars")
	subscriptions := router.Group("/subscriptions")
	{
		authentication.POST("/apple/login", controllers.AppleLogin)
		authentication.GET("/google/login", controllers.GoogleLogin)
		authentication.GET("/google")
		authentication.GET("/discord/login")
		authentication.GET("/discord")
		authentication.GET("/logout", controllers.Logout)
	}
	{
		users.GET("/@me", controllers.GetMe)
		users.PUT("/@me", controllers.UpdateUser)
	}
	{
		events.POST("/", controllers.CreateEvent)
		events.GET("/event_id", controllers.FetchEvent)
		events.PATCH("/", controllers.UpdateEvent)
		// POST, GET, DELETE, PUT-all fields update, PATCH-certain selected fields update
	}
	{
		groups.POST("/", controllers.CreateGroup)
		groups.GET("/", controllers.FetchGroup)
		groups.PATCH("/", controllers.UpdateGroup)
		// POST, GET, DELETE, PUT-all fields update, PATCH-certain selected fields update
	}
	{
		calendars.GET("/:calendar_id", controllers.FetchCalendar)
	}
	{
		subscriptions.POST("/", controllers.CreateSubscription)
		subscriptions.DELETE("/:calendar_id", controllers.DeleteSubscription)
		subscriptions.GET("/", controllers.FetchSubscriptions)
		// POST, GET, DELETE, PUT-all fields update, PATCH-certain selected fields update
	}
}

func cleanup(server *http.Server) {
	log.Println("shutdown signal received, cleaning up resources...")

	util.SaveCache(util.Nonces, "nonces")
	util.SaveCache(util.DailyUsers, "daily")
	util.SaveCache(util.ActiveUsers, "active")

	if err := server.Close(); err != nil {
		log.Println("server shutdown failed:", err)
	} else {
		log.Println("server shutdown successfully")
	}
}
