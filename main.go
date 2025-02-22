package main

import (
	"calenduh-backend/internal/controllers"
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/util"
	// "calenduh-backend/internal/handlers"
	"fmt"
	"github.com/joho/godotenv"
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

	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

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
	_ = router.Group("/event")
	_ = router.Group("/groups")
	calendars := router.Group("/calendars")
	_ = router.Group("/subscriptions")
	_ = router.Group("/groups/:group_id/members")
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
		users.GET("/@me", controllers.LoggedIn(controllers.GetMe))
		//users.POST("/", controllers.CreateUser) // Create a new user
		//users.GET("/:user_id", controllers.GetUser) // Get a specific user
		//users.PUT("/:user_id", controllers.UpdateUser) // Update user details
		//users.DELETE("/:user_id", controllers.DeleteUser) // Delete a user
	}
	{
		//events.POST("/:calendar_id", controllers.CreateEvent) // Create a new event
		//events.GET("/:event_id", controllers.GetEvent)        // Get a specific event
		//events.PUT("/:event_id", controllers.UpdateEvent)     // Update an event
		//events.DELETE("/:event_id", controllers.DeleteEvent)  // Delete an event
	}
	{
		//groups.GET("/", controllers.GetAllGroups) // List all groups
		//groups.POST("/", controllers.CreateGroup) // Create a new group
		//groups.GET("/:group_id", controllers.GetGroup) // Get a specific group
		//groups.PUT("/:group_id", controllers.UpdateGroup) // Update a group
		//groups.DELETE("/:group_id", controllers.DeleteGroup) // Delete a group
	}
	{
		//calendars.GET("/", controllers.GetAllCalendars)               // List all calendars
		//calendars.POST("/", controllers.CreateCalendar)               // Create a new calendar
		calendars.GET("/:calendar_id", controllers.FetchCalendar) // Get a specific calendar
		//calendars.PUT("/:calendar_id", controllers.UpdateCalendar)    // Update a calendar
		//calendars.DELETE("/:calendar_id", controllers.DeleteCalendar) // Delete a calendar
	}
	{
		//subscriptions.GET("/", controllers.GetAllSubscriptions)                        // List all subscriptions
		//subscriptions.POST("/", controllers.CreateSubscription)                        // Create a new subscription
		//subscriptions.GET("/:user_id/:calendar_id", controllers.GetSubscription)       // Get a specific subscription
		//subscriptions.DELETE("/:user_id/:calendar_id", controllers.DeleteSubscription) // Delete a subscription
	}

	{
		//groupMembers.GET("/", controllers.GetGroupMembers)              // List members of a group
		//groupMembers.POST("/", controllers.AddGroupMember)              // Add a member to a group
		//groupMembers.DELETE("/:user_id", controllers.RemoveGroupMember) // Remove a member from a group
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
