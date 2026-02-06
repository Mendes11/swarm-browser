package config

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	fetchTimeout  = 10 * time.Second
	cacheFileName = "clusters-cache.yml"
)

// FetchClustersFromURL fetches the clusters YAML from the given URL.
// On success, the response is cached to ~/.config/swarm-browser/clusters-cache.yml.
// On failure, falls back to the cached file if it exists.
// Returns the path to the YAML file to use.
func FetchClustersFromURL(url string) (string, error) {
	cacheDir, err := clustersCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine cache dir: %w", err)
	}
	cachePath := filepath.Join(cacheDir, cacheFileName)

	data, err := fetchURL(url)
	if err != nil {
		log.Printf("Warning: failed to fetch clusters from URL %s: %v", url, err)
		if _, statErr := os.Stat(cachePath); statErr == nil {
			log.Printf("Using cached clusters config from %s", cachePath)
			return cachePath, nil
		}
		return "", fmt.Errorf(
			"failed to fetch clusters from %s and no cached version available: %w",
			url, err,
		)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		log.Printf("Warning: failed to write clusters cache: %v", err)
	}

	return cachePath, nil
}

func fetchURL(url string) ([]byte, error) {
	client := &http.Client{Timeout: fetchTimeout}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return data, nil
}

func clustersCacheDir() (string, error) {
	return ConfigDir()
}
