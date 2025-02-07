package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/util"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// LocalLoginBody is the structure of a login request.
type LocalLoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AppleLoginBody struct {
	AuthorizationCode string  `json:"authorizationCode"`
	IdentityToken     string  `json:"identityToken"`
	UserId            string  `json:"user"`
	Email             *string `json:"email,omitempty"`
}

type AppleKeyResponse struct {
	Keys []AppleKey `json:"keys"`
}

type AppleKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
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

// AppleLogin
// @Summary Apple Login
// @Description Handles the login from Apple SignIn and creates a session.
func AppleLogin(c *gin.Context) {
	var appleLoginBody AppleLoginBody
	err := c.BindJSON(&appleLoginBody)
	if err != nil {
		message := gin.H{"message": "unable to parse body"}
		c.AbortWithStatusJSON(http.StatusBadRequest, message)
		return
	}

	log.Print(appleLoginBody.UserId)
	token, err := verifyToken(appleLoginBody.IdentityToken)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	} else {
		c.PureJSON(http.StatusOK, gin.H{"token": token})
	}
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

// Fetch Apple’s public keys and return the key that matches the given `kid`
func getSigningKey(kid string) (*rsa.PublicKey, error) {
	resp, err := http.Get(util.GetEnv("APPLE_AUTH_KEYS_URL"))
	if err != nil {
		return nil, fmt.Errorf("could not contact Apple endpoint: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Apple keys, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var appleKeyResponse AppleKeyResponse
	if err = json.Unmarshal(body, &appleKeyResponse); err != nil {
		return nil, err
	}

	for _, key := range appleKeyResponse.Keys {
		if key.Kid == kid {
			return convertJWKToPublicKey(key)
		}
	}

	return nil, errors.New("key not found")
}

func convertJWKToPublicKey(jwk AppleKey) (*rsa.PublicKey, error) {
	// Decode Base64 values
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to RSA Public Key
	pubKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}

	return pubKey, nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure token is signed with RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Fetch the signing key
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid not found in token header")
		}

		return getSigningKey(kid)
	}, jwt.WithValidMethods([]string{"RS256"}))

	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	return parsedToken, nil
}
