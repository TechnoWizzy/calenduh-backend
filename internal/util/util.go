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
func CreateNonce(clientId string) string {
	nonce := GetHash(clientId + time.Now().String())
	Nonces.Set(clientId, nonce, cache.DefaultExpiration)
	return nonce
}

// ValidateNonce determines whether a provided nonce is valid for the login requests.
func ValidateNonce(ip string, nonce string) bool {
	cachedNonce, found := Nonces.Get(ip)
	if !found {
		return false
	}
	if cachedNonce != nonce {
		return false
	}
	Nonces.Delete(ip)
	return true
}

func GetClientId(c *gin.Context) string {
	clientId, err := c.Cookie("client_id")
	if err != nil {
		clientId, err = gonanoid.New()
		if err != nil {
			clientId = uuid.New().String()
		}
		c.SetCookie("client_id", clientId, 1<<31-1, "/", c.Request.URL.String(), true, true)
	}

	DailyUsers.Set(clientId, true, TimeToMidnightEST())
	ActiveUsers.Set(clientId, true, cache.DefaultExpiration)
	return clientId
}

func GetHostAddress(c *gin.Context) string {
	host := c.GetHeader("Host")
	if host == "" {
		return c.Request.Host
	}
	return host
}

func GetProtocol(c *gin.Context) string {
	protocol := c.GetHeader("X-Protocol")
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
