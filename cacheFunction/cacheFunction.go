package cacheFunction

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Cache to store responses
var (
	cache        = make(map[string][]byte) // Map to store cached responses
	cacheMutex   = sync.RWMutex{}          // Mutex to handle concurrent access to the cache
	maxCacheSize = 3 * 1024 * 1024 * 1024  // Maximum cache size set to 3GB
	cacheDir     = "/var/cache/hammy/"     // Default cache directory
)

func init() {
	// Check if the primary cache directory is accessible
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		// Attempt to create the primary cache directory
		if err := os.Mkdir(cacheDir, os.ModePerm); err != nil {
			log.Printf("Failed to create primary cache directory: %v", err)
		}
		// Check if the fallback cache directory is accessible
		if _, err := os.Stat("./cache/"); os.IsNotExist(err) {
			// Attempt to create the fallback cache directory
			if err := os.Mkdir("./cache/", os.ModePerm); err != nil {
				log.Printf("DANGER: Failed to create fallback cache directory: %v", err)
			}
		}
		// Fallback to the project folder cache directory
		cacheDir = "./cache/"
		if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
			// Create the cache directory if it doesn't exist
			if err := os.Mkdir(cacheDir, os.ModePerm); err != nil {
				log.Fatalf("Failed to create cache directory: %v", err)
			}
		}
	}
}

// GetFromCache retrieves data from the cache
func GetFromCache(key string) ([]byte, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	data, exists := cache[key]
	return data, exists
}

// AddToCache adds data to the cache
func AddToCache(key string, data []byte) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if len(cache) >= maxCacheSize {
		log.Println("Cache is full, consider increasing the cache size or implementing a cache eviction policy")
		return
	}
	cache[key] = data
}

// SaveCacheToDisk saves the current cache to disk
func SaveCacheToDisk() error {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	for key, data := range cache {
		filePath := filepath.Join(cacheDir, "hammy-cache-"+key)
		if err := os.WriteFile(filePath, data, os.ModePerm); err != nil {
			log.Printf("Failed to write cache file %s: %v", filePath, err)
			continue
		}
	}
	return nil
}

// LoadCacheFromDisk loads the cache from disk
func LoadCacheFromDisk() error {
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		log.Printf("Failed to read cache directory: %v", err)
		return err
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(cacheDir, file.Name()))
		if err != nil {
			log.Printf("Failed to read cache file %s: %v", file.Name(), err)
			continue
		}
		cache[file.Name()] = data
	}
	return nil
}
