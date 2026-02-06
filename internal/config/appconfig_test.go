package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppConfig_FileNotFound(t *testing.T) {
	cfg, err := loadAppConfigFrom(t.TempDir())
	if err != nil {
		t.Fatalf("Expected no error for missing config, got: %v", err)
	}
	if cfg.ClustersFile != "" || cfg.ClustersURL != "" {
		t.Errorf("Expected zero-value AppConfig, got: %+v", cfg)
	}
}

func TestLoadAppConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	content := `clusters_file: /path/to/clusters.yml
clusters_url: https://example.com/clusters.yml
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	cfg, err := loadAppConfigFrom(dir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cfg.ClustersFile != "/path/to/clusters.yml" {
		t.Errorf("Expected clusters_file '/path/to/clusters.yml', got '%s'", cfg.ClustersFile)
	}
	if cfg.ClustersURL != "https://example.com/clusters.yml" {
		t.Errorf("Expected clusters_url 'https://example.com/clusters.yml', got '%s'", cfg.ClustersURL)
	}
}

func TestLoadAppConfig_PartialConfig(t *testing.T) {
	dir := t.TempDir()
	content := `clusters_url: https://example.com/clusters.yml
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	cfg, err := loadAppConfigFrom(dir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cfg.ClustersFile != "" {
		t.Errorf("Expected empty clusters_file, got '%s'", cfg.ClustersFile)
	}
	if cfg.ClustersURL != "https://example.com/clusters.yml" {
		t.Errorf("Expected clusters_url 'https://example.com/clusters.yml', got '%s'", cfg.ClustersURL)
	}
}

func TestLoadAppConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	content := `invalid: yaml: [broken`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	_, err := loadAppConfigFrom(dir)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}
