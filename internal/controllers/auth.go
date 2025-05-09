package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"calenduh-backend/internal/util"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"io"
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

// DiscordUser represents the structure return by the Discord User API.
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	GlobalName    string `json:"global_name"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
}

// Authorize is middleware that checks the login status of the current request.
// If a user is on an active session their ID is attached to the request under user_id.
func Authorize(c *gin.Context) {
	sessionId, err := c.Cookie("sessionId")
	if err != nil { // Check Auth
		sessionId = c.GetHeader("Authorization")
		if sessionId == "" {
			c.Next()
			return
		}
	}

	session, err := database.Db.Queries.GetSessionById(c, sessionId)
	if err != nil {
		c.Next()
		return
	}

	user, err := database.Db.Queries.GetUserById(c, session.UserID)
	if err != nil {
		c.Next()
		return
	}

	c.Set("user", &user)
	c.Next()
	return
}

func LoggedIn(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "not logged in"})
			return
		}
	}()

	user := *ParseUser(c)
	groups, err := database.Db.Queries.GetGroupsByUserId(c, user.UserID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "unable to fetch groups: " + err.Error()})
		return
	}

	c.Set("groups", &groups)
}

// AppleLogin
// @Summary Apple Login
// @Description Handles the login from Apple SignIn and creates a session.
func AppleLogin(c *gin.Context) {
	if err := database.Transaction(c, func(queries *sqlc.Queries) error {
		var appleLoginBody AppleLoginBody
		err := c.BindJSON(&appleLoginBody)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err)
			return nil
		}

		token, err := verifyToken(appleLoginBody.IdentityToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err)
			return nil
		}

		var email string
		if appleLoginBody.Email != nil {
			email = *appleLoginBody.Email
		} else {
			email = token.Claims.(jwt.MapClaims)["email"].(string)
		}

		user, err := queries.GetUserById(c, appleLoginBody.UserId)

		if err != nil { // User does not exist yet
			if errors.Is(err, pgx.ErrNoRows) {
				username := strings.Split(email, "@")[0]

				user, err = database.Db.Queries.CreateUser(c, sqlc.CreateUserParams{
					UserID:   appleLoginBody.UserId,
					Email:    email,
					Username: username,
				})

				if err != nil {
					return err
				}
			} else { // Failed to fetch user
				return err
			}
		}

		session, err := queries.InsertSession(c, sqlc.InsertSessionParams{
			SessionID: gonanoid.Must(),
			UserID:    user.UserID,
			Type:      sqlc.SessionTypeAPPLE,
		})

		if err != nil { // Failed to create session
			return err
		}

		c.PureJSON(http.StatusOK, gin.H{
			"sessionId": session.SessionID,
		})
		return nil
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
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
	if err := database.Transaction(c, func(queries *sqlc.Queries) error {
		state := c.Query("state")
		code := c.Query("code")
		validated, redirectUri := util.ValidateNonce(state)
		if !validated {
			message := gin.H{"message": "invalid state"}
			c.AbortWithStatusJSON(http.StatusBadRequest, message)
			return nil
		}

		if code == "" {
			message := gin.H{"message": "invalid code"}
			c.AbortWithStatusJSON(http.StatusBadRequest, message)
			return nil
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
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			message := gin.H{"message": "invalid code"}
			c.AbortWithStatusJSON(http.StatusBadRequest, message)
			return nil
		}

		var tokenData TokenData
		err = json.Unmarshal(resp.Body(), &tokenData)
		if err != nil {
			return err
		}

		googleApiUrl := util.GetEnv("GOOGLE_API_URL")

		resp, err = client.R().
			SetHeaders(map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer " + tokenData.AccessToken,
			}).
			Get(googleApiUrl + "/oauth2/v3/userinfo")
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return errors.New("unable to retrieve user data")
		}

		var googleUser GoogleUser
		err = json.Unmarshal(resp.Body(), &googleUser)
		if err != nil {
			return err
		}

		user, err := database.Db.Queries.GetUserById(c, googleUser.ID)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) { // User does not exist yet
				username := strings.Split(googleUser.Email, "@")[0]

				user, err = database.Db.Queries.CreateUser(c, sqlc.CreateUserParams{
					UserID:   googleUser.ID,
					Email:    googleUser.Email,
					Username: username,
				})

				if err != nil {
					return err
				}
			} else { // Failed to fetch user
				return err
			}
		}

		session, err := database.Db.Queries.InsertSession(c, sqlc.InsertSessionParams{
			SessionID:    gonanoid.Must(),
			UserID:       user.UserID,
			Type:         sqlc.SessionTypeGOOGLE,
			AccessToken:  &tokenData.AccessToken,
			RefreshToken: &tokenData.RefreshToken,
			ExpiresOn:    time.Now().Add(time.Duration(tokenData.ExpiresIn) * time.Second),
		})

		if err != nil {
			return err
		}

		c.SetCookie("sessionId", session.SessionID, tokenData.ExpiresIn, "/", c.Request.Host, false, true)
		c.Redirect(http.StatusTemporaryRedirect, *redirectUri+"?state="+state+"&sessionId="+session.SessionID)
		return nil
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
}

// DiscordLogin
// @Summary Discord Login
func DiscordLogin(c *gin.Context) {
	state := c.Query("state")
	if state == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "missing state"})
		return
	}
	localRedirectUri := c.Query("redirect_uri")
	redirectUri := util.GetProtocol(c) + c.Request.Host + util.GetEnv("DISCORD_OAUTH_URI")
	redirectUrl := util.GetEnv("DISCORD_OAUTH_URL")
	clientId := util.GetEnv("DISCORD_CLIENT_ID")
	responseType := "code"
	scope := "identify email"
	accessType := "offline"
	prompt := "none"
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

// DiscordAuth
// @Summary Discord Auth
func DiscordAuth(c *gin.Context) {
	if err := database.Transaction(c, func(queries *sqlc.Queries) error {
		state := c.Query("state")
		code := c.Query("code")
		validated, redirectUri := util.ValidateNonce(state)
		if !validated {
			message := gin.H{"message": "invalid state"}
			c.AbortWithStatusJSON(http.StatusBadRequest, message)
			return nil
		}

		if code == "" {
			message := gin.H{"message": "invalid code"}
			c.AbortWithStatusJSON(http.StatusBadRequest, message)
			return nil
		}

		client := resty.New()
		discordTokenUrl := util.GetEnv("DISCORD_OAUTH_TOKEN_URL")

		resp, err := client.R().
			SetFormData(map[string]string{
				"client_id":     util.GetEnv("DISCORD_CLIENT_ID"),
				"client_secret": util.GetEnv("DISCORD_CLIENT_SECRET"),
				"redirect_uri":  util.GetProtocol(c) + c.Request.Host + util.GetEnv("DISCORD_OAUTH_URI"),
				"grant_type":    "authorization_code",
				"code":          code,
			}).
			Post(discordTokenUrl)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			message := gin.H{"message": "invalid code"}
			c.AbortWithStatusJSON(http.StatusBadRequest, message)
			return nil
		}

		var tokenData TokenData
		err = json.Unmarshal(resp.Body(), &tokenData)
		if err != nil {
			return err
		}

		discordApiUrl := util.GetEnv("DISCORD_API_URL")

		resp, err = client.R().
			SetHeaders(map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer " + tokenData.AccessToken,
			}).
			Get(discordApiUrl + "/users/@me")
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			return err
		}

		var discordUser DiscordUser
		err = json.Unmarshal(resp.Body(), &discordUser)
		if err != nil {
			return err
		}

		user, err := database.Db.Queries.GetUserById(c, discordUser.ID)

		if err != nil { // User does not exist yet
			if errors.Is(err, pgx.ErrNoRows) {

				user, err = database.Db.Queries.CreateUser(c, sqlc.CreateUserParams{
					UserID:   discordUser.ID,
					Email:    discordUser.Email,
					Username: discordUser.Username,
				})

				if err != nil {
					return err
				}
			} else { // Failed to fetch user
				return err
			}
		}

		session, err := database.Db.Queries.InsertSession(c, sqlc.InsertSessionParams{
			SessionID:    gonanoid.Must(),
			UserID:       user.UserID,
			Type:         sqlc.SessionTypeDISCORD,
			AccessToken:  &tokenData.AccessToken,
			RefreshToken: &tokenData.RefreshToken,
			ExpiresOn:    time.Now().Add(time.Duration(tokenData.ExpiresIn) * time.Second),
		})

		if err != nil {
			return err
		}

		c.SetCookie("sessionId", session.SessionID, tokenData.ExpiresIn, "/",
			c.Request.Host, false, true)

		c.Redirect(http.StatusTemporaryRedirect, *redirectUri+"?state="+state+"&sessionId="+session.SessionID)
		return nil
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
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

	err = database.Db.Queries.DeleteSession(c, sessionId)
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

func GetAllSessions(c *gin.Context) {
	sessions, err := database.Db.Queries.GetAllSessions(c)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(http.StatusOK, make([]sqlc.Session, 0))
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, sessions)
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
