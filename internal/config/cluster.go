package config

import (
	"os"

	"github.com/mendes11/swarm-browser/internal/core/models"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type ClustersConfig struct {
	Clusters map[string]models.Cluster `yaml:"clusters"`
}

func LoadClustersConfig(path string) (*ClustersConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read clusters config file")
	}

	var config ClustersConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, "failed to parse clusters config")
	}

	return &config, nil
}

func (c *ClustersConfig) GetCluster(name string) (*models.Cluster, bool) {
	cluster, exists := c.Clusters[name]
	return &cluster, exists
}

func (c *ClustersConfig) ListClusters() []string {
	names := make([]string, 0, len(c.Clusters))
	for name := range c.Clusters {
		names = append(names, name)
	}
	return names
}
