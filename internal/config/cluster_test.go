package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadClustersConfig(t *testing.T) {
	yamlContent := `clusters:
  prod:
    name: "Production"
    host: "manager-01.example.com"
    nodes:
      manager-01:
        host: "manager-01.example.com"
        hostname: "manager-01.example.com"
  staging:
    name: "Staging"
    host: "swarm-02.test.com"
    nodes:
      swarm-02:
        host: "swarm-02.test.com"
        hostname: "swarm-02.test.com"
      worker-01:
        host: "worker-01.test.com"
        hostname: "worker-01.test.com"`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "clusters.yaml")
	
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	config, err := LoadClustersConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load clusters config: %v", err)
	}

	if len(config.Clusters) != 2 {
		t.Errorf("Expected 2 clusters, got %d", len(config.Clusters))
	}

	prod, exists := config.GetCluster("prod")
	if !exists {
		t.Error("Expected to find 'prod' cluster")
	}
	if prod.Name != "Production" {
		t.Errorf("Expected cluster name 'Production', got '%s'", prod.Name)
	}
	if prod.Host != "manager-01.example.com" {
		t.Errorf("Expected host 'manager-01.example.com', got '%s'", prod.Host)
	}

	staging, exists := config.GetCluster("staging")
	if !exists {
		t.Error("Expected to find 'staging' cluster")
	}
	if len(staging.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in staging cluster, got %d", len(staging.Nodes))
	}

	node, found := staging.GetNodeByHostname("worker-01.test.com")
	if !found {
		t.Error("Expected to find node with hostname 'worker-01.test.com'")
	}
	if node.Host != "worker-01.test.com" {
		t.Errorf("Expected node host 'worker-01.test.com', got '%s'", node.Host)
	}

	clusters := config.ListClusters()
	if len(clusters) != 2 {
		t.Errorf("Expected 2 cluster names, got %d", len(clusters))
	}

	nodes := staging.ListNodes()
	if len(nodes) != 2 {
		t.Errorf("Expected 2 node names in staging, got %d", len(nodes))
	}
}