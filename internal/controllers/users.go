package controllers

import (
	"errors"
	"hp-backend/internal/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserResponse struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// GetMe
// @Summary Get details of the current user
// @Description Fetches the user data for the currently authenticated user.
func GetMe(c *gin.Context) {
	user, err := ParseUser(c)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	response := *MakeUserResponse(user)
	c.PureJSON(http.StatusOK, response)
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

	user, err = database.UpdateUser(c, user.Id, &options)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	response := *MakeUserResponse(user)
	c.PureJSON(http.StatusOK, response)
}

func ParseUser(c *gin.Context) (*database.User, error) {
	v, found := c.Get("user")
	if !found {
		return nil, errors.New("user not found")
	}

	user := v.(database.User)
	return &user, nil
}

func MakeUserResponse(user *database.User) *UserResponse {
	if user == nil {
		return nil
	}

	return &UserResponse{
		Id:       user.Id,
		Email:    user.Email,
		Username: user.Username,
	}
}

func ValidateUpdateUserOptions(options *database.UpdateUserOptions) error {
	if options.Email == "" {
		return errors.New("email cannot be empty")
	}
	if options.Username == "" {
		return errors.New("username cannot be empty")
	}
	if options.Password == "" {
		return errors.New("password cannot be empty")
	}

	return nil
}
