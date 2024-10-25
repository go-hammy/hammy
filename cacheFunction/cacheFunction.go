package cacheFunction

import (
	"log"
	"os"
	"path/filepath"
	"strings"
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
	ensureCacheDirectory()
}

// ensureCacheDirectory checks and creates the necessary cache directories
func ensureCacheDirectory() {
	if !isDirectoryAccessible(cacheDir) {
		if err := os.Mkdir(cacheDir, os.ModePerm); err != nil {
			log.Printf("Failed to create primary cache directory: %v", err)
		}
		if !isDirectoryAccessible("./cache/") {
			if err := os.Mkdir("./cache/", os.ModePerm); err != nil {
				log.Printf("DANGER: Failed to create fallback cache directory: %v", err)
			}
		}
		cacheDir = "./cache/"
		if !isDirectoryAccessible(cacheDir) {
			if err := os.Mkdir(cacheDir, os.ModePerm); err != nil {
				log.Fatalf("Failed to create cache directory: %v", err)
			}
		}
	}
}

// isDirectoryAccessible checks if a directory is accessible
func isDirectoryAccessible(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
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
		filePath := filepath.Join(cacheDir, sanitizeFileName(key))
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

// sanitizeFileName ensures the file name is safe for use in the file system
func sanitizeFileName(name string) string {
	return strings.ReplaceAll(name, "/", "_")
}
