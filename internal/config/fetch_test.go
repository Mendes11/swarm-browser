package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchClustersFromURL_Success(t *testing.T) {
	yamlContent := `clusters:
  test:
    name: "Test"
    host: "localhost"
    nodes:
      node1:
        host: "localhost"
        hostname: "node1"`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(yamlContent))
	}))
	defer server.Close()

	path, err := FetchClustersFromURL(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read returned path: %v", err)
	}

	if string(data) != yamlContent {
		t.Errorf("Content mismatch.\nExpected: %q\nGot:      %q", yamlContent, string(data))
	}
}

func TestFetchClustersFromURL_FallbackToCache(t *testing.T) {
	cacheDir, err := clustersCacheDir()
	if err != nil {
		t.Fatalf("Failed to get cache dir: %v", err)
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	cachePath := filepath.Join(cacheDir, cacheFileName)
	cachedContent := `clusters:
  cached:
    name: "Cached"
    host: "cached-host"
    nodes: {}`
	if err := os.WriteFile(cachePath, []byte(cachedContent), 0644); err != nil {
		t.Fatalf("Failed to write cache file: %v", err)
	}
	defer os.Remove(cachePath)

	// Use a URL that will fail to connect
	path, err := FetchClustersFromURL("http://127.0.0.1:1/nonexistent")
	if err != nil {
		t.Fatalf("Expected fallback to cache, got error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read cached path: %v", err)
	}
	if string(data) != cachedContent {
		t.Errorf("Expected cached content, got %q", string(data))
	}
}

func TestFetchClustersFromURL_NoCacheFails(t *testing.T) {
	// Ensure no cache file exists
	cacheDir, err := clustersCacheDir()
	if err != nil {
		t.Fatalf("Failed to get cache dir: %v", err)
	}
	os.Remove(filepath.Join(cacheDir, cacheFileName))

	_, err = FetchClustersFromURL("http://127.0.0.1:1/nonexistent")
	if err == nil {
		t.Fatal("Expected error when URL fails and no cache exists")
	}
}

func TestFetchClustersFromURL_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Ensure no cache so we get an error
	cacheDir, _ := clustersCacheDir()
	os.Remove(filepath.Join(cacheDir, cacheFileName))

	_, err := FetchClustersFromURL(server.URL)
	if err == nil {
		t.Fatal("Expected error for 500 response with no cache")
	}
}
