package main

import (
	"calenduh-backend/internal/controllers"
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/util"
	"fmt"
	"github.com/gin-contrib/cors"
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

	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
	}

	// Database Connection

	if err := database.New(util.GetEnv("POSTGRESQL_URL")); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if database.Db.Pool != nil {
			database.Db.Pool.Close()
		}
		if database.Db.Conn != nil {
			if err := database.Db.Conn.Close(); err != nil {
				log.Println("failed to close database connection:", err)
			}
		}
	}()

	if env := util.GetEnv("GO_ENV"); env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8081"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	config.AllowCredentials = true
	// config.AllowOrigins = []string{"http://google.com", "http://facebook.com"}
	// config.AllowAllOrigins = true

	// Router setup
	router := gin.Default()
	router.Use(CORSMiddleware())
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
}

func setupRoutes(router *gin.Engine) {
	authentication := router.Group("/auth")
	files := router.Group("/files")
	users := router.Group("/users")
	events := router.Group("/events")
	groups := router.Group("/groups")
	calendars := router.Group("/calendars")
	subscriptions := router.Group("/subscriptions")
	_ = router.Group("/groups/:group_id/members")
	{ // Auth
		authentication.POST("/apple/login", controllers.AppleLogin)
		authentication.GET("/google/login", controllers.GoogleLogin)
		authentication.GET("/google", controllers.GoogleAuth)
		authentication.GET("/discord/login", controllers.DiscordLogin)
		authentication.GET("/discord", controllers.DiscordAuth)
		authentication.GET("/logout", controllers.Logout)
		authentication.GET("/sessions", controllers.GetAllSessions)
	}
	{ // Files
		// files.POST("/:key", controllers.LoggedIn, controllers.UploadFile)   // Upload Profile Picture
		// files.DELETE("/:key", controllers.LoggedIn, controllers.DeleteFile) // Delete Profile Picture
		files.POST("/upload", controllers.LoggedIn, controllers.UploadFile)
		files.POST("/uploadFile", controllers.LoggedIn, controllers.UploadFileNotAProfilePicture)
		files.PUT("/profile", controllers.LoggedIn, controllers.UpdateProfilePicture) // note: unused I think
		files.DELETE("/profile", controllers.LoggedIn, controllers.DeleteProfilePicture)
		files.GET("/profile/url", controllers.LoggedIn, controllers.GetProfilePictureURL)
		files.POST("/uploadEventImage/:calendar_id/:event_id", controllers.LoggedIn, controllers.CreateEventImage)
		// files.PUT("/updateEventImage/:calendar_id/:event_id", controllers.LoggedIn, controllers.UpdateEventImage)
		files.DELETE("/deleteEventImage/:calendar_id/:event_id", controllers.LoggedIn, controllers.DeleteEventImage)
	}
	{ // Users
		users.GET("/", controllers.GetAllUsers)                                       // Get all users
		users.GET("/@me", controllers.LoggedIn, controllers.GetMe)                    // Get self user
		users.GET("/:user_id", controllers.LoggedIn, controllers.GetUser)             // Get a specific user
		users.PUT("/:user_id", controllers.LoggedIn, controllers.UpdateUser)          // Update user details
		users.POST("/@local", controllers.LoggedIn, controllers.UploadLocalCalendars) // Upload local user calendars and events
		users.DELETE("/@me", controllers.LoggedIn, controllers.DeleteMe)              // Delete self user
		users.DELETE("/@all", controllers.DeleteAllUsers)                             // Delete all users
		users.DELETE("/:user_id", controllers.DeleteUser)                             // Delete user by id
	}
	{ // Events
		events.GET("/", controllers.WithRange, controllers.GetAllEvents)                                         // List all events
		events.GET("/@me", controllers.WithRange, controllers.LoggedIn, controllers.GetUserEvents)               // Get all events for a user that start today
		events.GET("/:calendar_id", controllers.WithRange, controllers.LoggedIn, controllers.GetCalendarEvents)  // Get Calendar events
		events.GET("/:calendar_id/:event_id", controllers.WithRange, controllers.LoggedIn, controllers.GetEvent) // Get a specific event
		events.POST("/:calendar_id", controllers.LoggedIn, controllers.CreateEvent)                              // Create a new event
		events.PUT("/:calendar_id/:event_id", controllers.LoggedIn, controllers.UpdateEvent)                     // Update an event
		events.DELETE("/@all", controllers.DeleteAllEvents)                                                      // Delete all events
		events.DELETE("/@prune", controllers.LoggedIn, controllers.PruneEvents)                                  // Prune events that are no longer occurring
		events.DELETE("/:calendar_id/:event_id", controllers.LoggedIn, controllers.DeleteEvent)                  // Delete an event
	}
	{ // Groups
		groups.GET("/", controllers.GetAllGroups)                                      // List all groups
		groups.GET("/@me", controllers.LoggedIn, controllers.GetMyGroups)              // List all user groups
		groups.GET("/:group_id", controllers.LoggedIn, controllers.GetGroup)           // Get a specific group
		groups.POST("/join/:invite_code", controllers.LoggedIn, controllers.JoinGroup) // Join a group by code
		groups.POST("/leave/:group_id", controllers.LoggedIn, controllers.LeaveGroup)  // Leave a group
		groups.POST("/", controllers.LoggedIn, controllers.CreateGroup)                // Create a new group
		groups.PUT("/:group_id", controllers.LoggedIn, controllers.UpdateGroup)        // Update a group
		groups.DELETE("/:group_id", controllers.LoggedIn, controllers.DeleteGroup)     // Delete a group
	}
	{ // Calendars
		calendars.GET("/", controllers.GetAllCalendars)                                          // List all calendars
		calendars.GET("/@me", controllers.LoggedIn, controllers.GetUserCalendars)                // List all calendars owned by user
		calendars.GET("/@groups", controllers.LoggedIn, controllers.GetAllGroupCalendars)        // List all calendars owned by user groups
		calendars.GET("/@public", controllers.LoggedIn, controllers.GetAllPublicCalendars)       // List all public calendars
		calendars.GET("/@groups/:group_id", controllers.LoggedIn, controllers.GetGroupCalendars) // List all calendars owned by a single user group
		calendars.GET("/@subscribed", controllers.LoggedIn, controllers.GetSubscribedCalendars)  // List all the calendars subscribed to by user
		calendars.GET("/:calendar_id", controllers.GetCalendar)                                  // Get a specific calendar
		calendars.POST("/", controllers.LoggedIn, controllers.CreateUserCalendar)                // Create a new user calendar
		calendars.POST("/:group_id", controllers.LoggedIn, controllers.CreateGroupCalendar)      // Create a new group calendar
		calendars.PUT("/:calendar_id", controllers.LoggedIn, controllers.UpdateCalendar)         // Update a calendar
		calendars.DELETE("/@all", controllers.DeleteAllCalendars)                                // Delete all calendars
		calendars.DELETE("/:calendar_id", controllers.LoggedIn, controllers.DeleteCalendar)      // Delete a calendar
	}
	{ // Subscriptions
		subscriptions.GET("/", controllers.GetAllSubscriptions) // List all subscriptions
		subscriptions.POST("/", controllers.CreateSubscription) // Create a new subscription
		//subscriptions.GET("/:user_id/:calendar_id", controllers.GetSubscription) // Get a specific subscription
		subscriptions.DELETE("/:calendar_id", controllers.DeleteMySubscription)        // Delete a subscription
		subscriptions.DELETE("/:calendar_id/:user_id", controllers.DeleteSubscription) // Delete a subscription
	}
	{ // GroupMembers
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

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
