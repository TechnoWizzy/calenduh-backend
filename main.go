package main

import (
	"calenduh-backend/internal/controllers"
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/util"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Do not remove
)

func main() {
	timeStarted := time.Now()

	// Database Connection
	instance, err := database.New(util.GetEnv("POSTGRESQL_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if instance.Pool != nil {
			instance.Pool.Close()
		}
		if instance.Conn != nil {
			if err := instance.Conn.Close(); err != nil {
				log.Println("failed to close database connection:", err)
			}
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
	//events := router.Group("/event")
	//groups := router.Group("/groups")
	calendars := router.Group("/calendars")
	//subscriptions := router.Group("/subscriptions")
	{
		authentication.POST("/apple/login", controllers.AppleLogin)
		authentication.GET("/google/login", controllers.GoogleLogin)
		authentication.GET("/google", controllers.GoogleAuth)
		authentication.GET("/discord/login", controllers.DiscordLogin)
		authentication.GET("/discord", controllers.DiscordAuth)
		authentication.GET("/logout", controllers.Logout)
		authentication.GET("/sessions", controllers.GetAllSessions)
	}
	{
		users.GET("/", controllers.GetAllUsers)
		users.GET("/@me", controllers.GetMe)
	}
	{
		// POST, GET, DELETE, PUT-all fields update, PATCH-certain selected fields update
	}
	{
		// POST, GET, DELETE, PUT-all fields update, PATCH-certain selected fields update
	}
	{
		calendars.GET("/:calendar_id", controllers.FetchCalendar)
	}
	{
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
