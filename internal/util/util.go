package util

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/matoous/go-nanoid/v2"
	"github.com/patrickmn/go-cache"
	"log"
	"net/http"
	"os"
	"time"
)

var globalRateLimitStore = ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
	Rate:  time.Second,
	Limit: 1 << 3,
})

var loginRateLimitStore = ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
	Rate:  time.Minute,
	Limit: 1 << 3,
})

var GlobalRateLimit = ratelimit.RateLimiter(globalRateLimitStore, &ratelimit.Options{
	ErrorHandler: HandleRateLimit,
	KeyFunc:      GetClientId,
})

var LoginRateLimit = ratelimit.RateLimiter(loginRateLimitStore, &ratelimit.Options{
	ErrorHandler: HandleRateLimit,
	KeyFunc:      GetClientId,
})

func GetEnv(key string) string {
	value, found := os.LookupEnv(key)
	if !found {
		log.Fatal("Could not load environment variable: " + key)
	}
	return value
}

func GetHash(value string) string {
	hash := sha256.New()
	hash.Write([]byte(value))
	return hex.EncodeToString(hash.Sum(nil))
}

// CreateNonce creates an ID and time-based nonce for login requests.
func CreateNonce(state string, redirectUri string) string {
	Nonces.Set(state, redirectUri, cache.DefaultExpiration)
	return state
}

// ValidateNonce determines whether a provided nonce is valid for the login requests.
func ValidateNonce(code string) (bool, *string) {
	v, found := Nonces.Get(code)
	if !found {
		return false, nil
	}
	Nonces.Delete(code)
	redirectUri := v.(string)
	return true, &redirectUri
}

func GetClientId(c *gin.Context) string {
	log.Println("Host: " + c.Request.Host)
	clientId, err := c.Cookie("client_id")
	if err != nil {
		clientId, err = gonanoid.New()
		if err != nil {
			clientId = uuid.New().String()
		}
		c.SetCookie("client_id", clientId, 1<<32-1, "/", c.Request.Host, true, true)
	}

	DailyUsers.Set(clientId, true, TimeToMidnightEST())
	ActiveUsers.Set(clientId, true, cache.DefaultExpiration)
	return clientId
}

func GetProtocol(c *gin.Context) string {
	protocol := c.GetHeader("X-Forwarded-Proto")
	if protocol == "" {
		return "http://"
	} else {
		return "https://"
	}
}

func HandleRateLimit(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).Round(time.Second).String())
}

func TimeToMidnightEST() time.Duration {
	location, _ := time.LoadLocation("EST")
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, location)
	return time.Until(midnight)
}
