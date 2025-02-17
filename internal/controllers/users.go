package controllers

import (
	"calenduh-backend/internal/database"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetMe
// @Summary Get details of the current user
// @Description Fetches the user data for the currently authenticated user.
func GetMe(c *gin.Context) {
	user, err := ParseUser(c)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.PureJSON(http.StatusOK, user)
	return
}

// UpdateUser
// @Summary Update an existing user
// @Description Updates user details by user ID. Requires admin privileges.
func UpdateUser(c *gin.Context) {
	user, err := ParseUser(c)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var options database.UpdateUserOptions
	if err := c.BindJSON(&options); err != nil {
		message := gin.H{"message": "body could not be read"}
		c.AbortWithStatusJSON(http.StatusBadRequest, message)
		return
	}

	if err := ValidateUpdateUserOptions(&options); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	user, err = database.UpdateUser(c, user.UserId, &options)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	c.PureJSON(http.StatusOK, user)
}

func ParseUser(c *gin.Context) (*database.User, error) {
	v, found := c.Get("user")
	if !found {
		return nil, errors.New("user not found")
	}

	user := v.(database.User)
	return &user, nil
}

func ValidateUpdateUserOptions(options *database.UpdateUserOptions) error {
	if options.Email == "" {
		return errors.New("email cannot be empty")
	}
	if options.Username == "" {
		return errors.New("username cannot be empty")
	}

	return nil
}
