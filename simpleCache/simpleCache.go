package simpleCache

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

//ExchangeCache Holds the information for the cache
type ExchangeCache struct {
	Rates   map[string]float64 `json:"rates"`
	Base    string             `json:"base"`
	Expires int64              `json:"expires"`
}

const cacheFile = "cache.json"

//GetCache loads the cache from file if there is one or returns an error if there isn't
func GetCache() (ExchangeCache, error) {
	var cache ExchangeCache
	if cacheExists() {
		configFile, err := os.Open(cacheFile)
		defer configFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}

		jsonParser := json.NewDecoder(configFile)
		jsonParser.Decode(&cache)

		//Check if the cache has expired
		if cache.Expires > time.Now().Unix() {
			return cache, errors.New("Cache Expired")
		}

		return cache, nil
	}
	return cache, errors.New("No cache")
}

//SaveCache saves a cache into a file
func SaveCache(cache ExchangeCache) {
	if !cacheExists() {
		createCache()
	}
	f, err := os.OpenFile(cacheFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	jsonBytes, _ := json.Marshal(cache)
	f.Write(jsonBytes)
	f.Close()
}

func cacheExists() bool {
	if _, err := os.Stat(cacheFile); err == nil {
		return true
	}
	return false
}

func createCache() {
	_, err := os.Create(cacheFile)
	if err != nil {
		panic(err)
	}
}
