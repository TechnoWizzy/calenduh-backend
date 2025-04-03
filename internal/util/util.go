package util

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"log"
	"net/http"
	"os"
	"time"
)

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

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
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
