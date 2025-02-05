package controllers

import (
	"hp-backend/internal/database"
	"hp-backend/internal/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// LocalLoginBody is the structure of a login request.
type LocalLoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Authorize is middleware that checks the login status of the current request.
// If a user is on an active session their ID is attached to the request under user_id.
func Authorize(c *gin.Context) {
	sessionId, err := c.Cookie("session_id")
	if err != nil { // No session
		c.Next()
		return
	}

	session, err := database.FetchSessionById(c, sessionId)
	if err != nil {
		c.Next()
		return
	}

	user, err := database.FetchUserById(c, session.UserId)
	if err != nil {
		c.Next()
		return
	}

	c.Set("user", *user)
	c.Next()
	return
}

// LocalLogin
// @Summary Local login
// @Description Handles user login via email and password with rate limiting.
func LocalLogin(c *gin.Context) {
	var localLoginBody LocalLoginBody
	err := c.BindJSON(&localLoginBody)
	if err != nil {
		message := gin.H{"message": "unable to parse body"}
		c.AbortWithStatusJSON(http.StatusBadRequest, message)
		return
	}

	user, err := database.FetchUserByEmail(c, localLoginBody.Email)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	hash := util.GetHash(localLoginBody.Password + util.GetHash(localLoginBody.Email))
	if hash != user.Password {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	expireTime := 60 * 60 * 24
	session, err := database.CreateSession(c, &database.CreateSessionOptions{
		UserId:       user.Id,
		AccessToken:  localLoginBody.Password,
		Type:         database.LocalSession,
		RefreshToken: "",
		ExpiresOn:    time.Now().Add(time.Duration(expireTime) * time.Second),
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.SetCookie("session_id", session.Id, expireTime, "/", c.Request.Host, true, true)
	c.Redirect(http.StatusTemporaryRedirect, util.GetProtocol(c)+util.GetHostAddress(c)+"/users/@me")
	return
}

// Logout
// @Summary Logout
// @Description Logs the user out by deleting the session cookie and session data.
func Logout(c *gin.Context) {
	sessionId, err := c.Cookie("session_id")
	if err != nil { // No session
		return
	}

	err = database.DeleteSession(c, sessionId)
	if err != nil {
		message := gin.H{
			"message": "unable to execute query: DeleteSession",
			"error":   err.Error(),
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, message)
		return
	}

	c.SetCookie("session_id", "", 0, "/", c.Request.Host, true, true)
	c.Status(http.StatusOK)
	return
}
