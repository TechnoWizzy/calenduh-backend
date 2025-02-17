package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/util"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AppleLoginBody struct {
	AuthorizationCode string  `json:"authorizationCode"`
	IdentityToken     string  `json:"identityToken"`
	UserId            string  `json:"user"`
	Email             *string `json:"email,omitempty"`
}

// TokenData represents the structure of a successful OAuth2 flow.
// AccessToken is used to make requests to its respective API to retrieve user information.
// RefreshToken is used to obtain a new AccessToken once ExpiresIn seconds have elapsed.
// Scope determines what information can be obtained from the API about the user.
type TokenData struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// GoogleUser represents the structure returned by Google Persons API.
type GoogleUser struct {
	ID       string `json:"sub"`
	Picture  string `json:"picture"`
	Email    string `json:"email"`
	Verified bool   `json:"email_verified"`
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

	spew.Dump(appleLoginBody)
	token, err := verifyToken(appleLoginBody.IdentityToken)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var email string
	if appleLoginBody.Email != nil {
		email = *appleLoginBody.Email
	} else {
		email = token.Claims.(jwt.MapClaims)["email"].(string)
	}

	user, err := database.FetchUserById(c, appleLoginBody.UserId)

	if err != nil { // No user
		if errors.Is(err, mongo.ErrNoDocuments) {
			user, err = database.CreateUser(c, &database.CreateUserOptions{
				Id:       &appleLoginBody.UserId,
				Email:    email,
				Username: email,
			})

			if err != nil { // Could not create user
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
				return
			}
		} else { // Failed to fetch user
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}

	session, err := database.CreateSession(c, &database.CreateSessionOptions{
		UserId: user.UserId,
		Type:   database.AppleSession,
	})

	if err != nil { // Failed to create session
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.PureJSON(http.StatusOK, gin.H{
		"session": session,
		"user":    user,
	})
	return
}

// GoogleLogin
// @Summary Google Login
func GoogleLogin(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "missing state"})
		return
	}
	localRedirectUri := c.Query("redirect_uri")
	redirectUri := util.GetProtocol(c) + c.Request.Host + util.GetEnv("GOOGLE_OAUTH_URI")
	redirectUrl := util.GetEnv("GOOGLE_OAUTH_URL")
	clientId := util.GetEnv("GOOGLE_CLIENT_ID")
	responseType := "code"
	scope := "https://www.googleapis.com/auth/userinfo.email"
	accessType := "offline"
	prompt := "select_account"
	state = util.CreateNonce(state, localRedirectUri)
	params := fmt.Sprintf(
		"?response_type=%s&client_id=%s&scope=%s&access_type=%s&prompt=%s&redirect_uri=%s&state=%s",
		responseType,
		clientId,
		url.QueryEscape(scope),
		accessType,
		prompt,
		url.QueryEscape(redirectUri),
		state)
	c.Redirect(http.StatusTemporaryRedirect, redirectUrl+params)
}

// GoogleAuth
// @Summary Google Auth
func GoogleAuth(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")
	validated, redirectUri := util.ValidateNonce(state)
	if !validated {
		message := gin.H{"message": "invalid state"}
		c.AbortWithStatusJSON(http.StatusBadRequest, message)
		return
	}

	if code == "" {
		message := gin.H{"message": "invalid code"}
		c.AbortWithStatusJSON(http.StatusBadRequest, message)
		return
	}

	client := resty.New()
	googleTokenUrl := util.GetEnv("GOOGLE_OAUTH_TOKEN_URL")

	resp, err := client.R().
		SetFormData(map[string]string{
			"client_id":     util.GetEnv("GOOGLE_CLIENT_ID"),
			"client_secret": util.GetEnv("GOOGLE_CLIENT_SECRET"),
			"redirect_uri":  util.GetProtocol(c) + c.Request.Host + util.GetEnv("GOOGLE_OAUTH_URI"),
			"grant_type":    "authorization_code",
			"code":          code,
		}).
		Post(googleTokenUrl)
	if err != nil {
		message := gin.H{
			"message": "failed to acquire oauth token",
			"error":   err.Error(),
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, message)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		message := gin.H{"message": "invalid code"}
		c.AbortWithStatusJSON(http.StatusBadRequest, message)
		return
	}

	var tokenData TokenData
	err = json.Unmarshal(resp.Body(), &tokenData)
	if err != nil {
		message := gin.H{
			"message": "unable to unmarshal tokenData body",
			"error":   err.Error(),
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, message)
		return
	}

	googleApiUrl := util.GetEnv("GOOGLE_API_URL")

	resp, err = client.R().
		SetHeaders(map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + tokenData.AccessToken,
		}).
		Get(googleApiUrl + "/oauth2/v3/userinfo")
	if err != nil {
		message := gin.H{
			"message": "unable to retrieve user data",
			"error":   err.Error(),
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, message)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		message := gin.H{
			"message": "unable to retrieve user data",
			"error":   resp.Body(),
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, message)
		return
	}

	var googleUser GoogleUser
	err = json.Unmarshal(resp.Body(), &googleUser)
	if err != nil {
		message := gin.H{
			"message": "unable to unmarshal googleUser body",
			"error":   err.Error(),
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, message)
		return
	}

	user, err := database.FetchUserByEmail(c, googleUser.Email)
	if err != nil { // User does not exist yet
		username := strings.Split(googleUser.Email, "@")[0]

		user, err = database.CreateUser(c, &database.CreateUserOptions{
			Email:    googleUser.Email,
			Username: username,
		})

		if err != nil {
			message := gin.H{
				"message": "unable to create new user",
				"error":   err.Error(),
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, message)
			return
		}
	}

	session, err := database.CreateSession(c, &database.CreateSessionOptions{
		UserId:       user.UserId,
		Type:         database.GoogleSession,
		AccessToken:  tokenData.AccessToken,
		RefreshToken: tokenData.RefreshToken,
		ExpiresOn:    time.Now().Add(time.Duration(tokenData.ExpiresIn) * time.Second),
	})

	if err != nil {
		message := gin.H{
			"message": "unable to execute query: CreateSession",
			"error":   err.Error(),
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, message)
		return
	}

	c.SetCookie("session_id", session.SessionId, tokenData.ExpiresIn, "/",
		c.Request.Host, false, true)

	log.Print(*redirectUri)
	c.Redirect(http.StatusTemporaryRedirect, *redirectUri+"?state="+state+"&session_id="+session.SessionId)
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
