package util

import (
	"bytes"
	"encoding/gob"
	"github.com/patrickmn/go-cache"
	"log"
	"os"
	"time"
)

var CachePath = "/var/lib/cache/"
var Nonces = CreateCache(5*time.Minute, time.Minute, "nonces")
var DailyUsers = CreateCache(24*time.Hour, time.Minute, "daily")
var ActiveUsers = CreateCache(15*time.Minute, time.Minute, "active")

func CreateCache(defaultExpiration time.Duration, cleanupInterval time.Duration, name string) *cache.Cache {
	file, err := os.ReadFile(CachePath + name)
	if err != nil {
		//log.Printf("Unable to read %s cache from file: \n%s", name, err.Error())
		return cache.New(defaultExpiration, cleanupInterval)
	}

	values := make(map[string]cache.Item)
	decoder := gob.NewDecoder(bytes.NewBuffer(file))
	if err := decoder.Decode(&values); err != nil {
		//log.Printf("Unable to decode %s cache: \n%s", name, err.Error())
		return cache.New(defaultExpiration, cleanupInterval)
	}

	return cache.NewFrom(defaultExpiration, cleanupInterval, values)
}

func SaveCache(c *cache.Cache, name string) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	if err := encoder.Encode(c.Items()); err != nil {
		log.Printf("Unable to save %s cache: \n%s", name, err.Error())
	}

	if err := os.WriteFile(CachePath+name, buf.Bytes(), os.ModePerm); err != nil {
		log.Printf("Unable to save %s cache: \n%s", name, err.Error())
	}

	log.Printf("Saved %s cache\n", name)

	return
}
