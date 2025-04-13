package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"net/http"
)

type UploadLocalCalendarsParams struct {
	Calendars []sqlc.CreateCalendarParams `json:"calendars"`
	Events    []sqlc.CreateEventParams    `json:"events"`
}

// GetMe
// @Summary Get details of the current user
// @Description Fetches the user data for the currently authenticated user.
func GetMe(c *gin.Context) {
	user := *ParseUser(c)
	c.PureJSON(http.StatusOK, user)
}

// GetAllUsers
// @Summary Lists all users in the database
// @Description Debug route to see all users currently existing
func GetAllUsers(c *gin.Context) {
	users, err := database.Db.Queries.GetAllUsers(c)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(http.StatusOK, make([]sqlc.User, 0))
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.PureJSON(http.StatusOK, users)
}

// GetUser
// @Summary Gets a User by ID
// @Description Used to retrieve user details for an account other than the one logged in
func GetUser(c *gin.Context) {
	userId := c.Param("user_id")
	if userId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	user, err := database.Db.Queries.GetUserById(c, userId)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.PureJSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	user := *ParseUser(c)

	var updateUserParams sqlc.UpdateUserParams
	if err := c.ShouldBindJSON(&updateUserParams); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateUserParams.UserID = user.UserID

	user, err := database.Db.Queries.UpdateUser(c, updateUserParams)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.PureJSON(http.StatusOK, user)
}

func DeleteMe(c *gin.Context) {
	user := *ParseUser(c)
	groups := *ParseGroups(c)
	if err := database.Transaction(c, func(queries *sqlc.Queries) error {
		for _, group := range groups {
			groupMembers, err := queries.GetGroupMembers(c, group.GroupID)
			if err != nil {
				return err
			}
			if len(groupMembers) == 1 {
				if err := queries.DeleteGroup(c, group.GroupID); err != nil {
					return err
				}
			}
		}

		if err := database.Db.Queries.DeleteUser(c, user.UserID); err != nil {
			switch {
			case errors.Is(err, pgx.ErrNoRows):
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			default:
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return nil
		}

		return nil
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.PureJSON(http.StatusOK, gin.H{"status": "deleted"})
}

func DeleteUser(c *gin.Context) {
	userId := c.Param("user_id")
	if err := database.Transaction(c, func(queries *sqlc.Queries) error {
		if userId == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return nil
		}

		groups, err := queries.GetGroupsByUserId(c, userId)
		if err != nil {
			switch {
			case errors.Is(err, pgx.ErrNoRows):
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return nil
			default:
				return err
			}
		}

		for _, group := range groups {
			groupMembers, err := queries.GetGroupMembers(c, group.GroupID)
			if err != nil {
				return err
			}
			if len(groupMembers) == 1 {
				if err := queries.DeleteGroup(c, group.GroupID); err != nil {
					return err
				}
			}
		}

		if err := database.Db.Queries.DeleteUser(c, userId); err != nil {
			switch {
			case errors.Is(err, pgx.ErrNoRows):
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			default:
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return nil
		}

		return nil
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.PureJSON(http.StatusOK, gin.H{"status": "deleted"})
}

func DeleteAllUsers(c *gin.Context) {
	err := database.Db.Queries.DeleteAllUsers(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "all users deleted successfully"})
}

func UploadLocalCalendars(c *gin.Context) {
	user := *ParseUser(c)
	var input UploadLocalCalendarsParams
	if err := c.BindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.Transaction(c, func(queries *sqlc.Queries) error {
		// Calendars first
		for _, createCalendarParams := range input.Calendars {
			createCalendarParams.UserID = &user.UserID
			if _, err := queries.CreateCalendar(c, createCalendarParams); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return err
			}
		}

		// Then Events
		for _, createEventParams := range input.Events {
			if _, err := queries.CreateEvent(c, createEventParams); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return err
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "UploadLocalCalendars successful"})
		return nil
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func ParseUser(c *gin.Context) *sqlc.User {
	v, found := c.Get("user")
	if !found {
		panic(errors.New("user not found"))
	}
	user, ok := v.(*sqlc.User)
	if !ok {
		panic(errors.New("user type assertion failed"))
	}
	return user
}
