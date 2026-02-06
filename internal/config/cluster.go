package config

import (
	"os"

	"github.com/mendes11/swarm-browser/internal/core/models"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type ClustersConfig struct {
	Clusters map[string]models.Cluster `yaml:"clusters" mapstructure:"clusters"`
	Commands map[string]models.Command `yaml:"commands,omitempty" mapstructure:"commands"`
}

// ResolveCommandsForCluster returns the effective commands for a given cluster.
// If the cluster defines its own commands, those are used (merged with global definitions).
// If the cluster has no commands, all global commands are returned.
func (c *ClustersConfig) ResolveCommandsForCluster(clusterName string) []models.Command {
	cluster, exists := c.Clusters[clusterName]
	if !exists {
		return nil
	}

	// If cluster has no commands section, return all global commands
	if len(cluster.Commands) == 0 {
		result := make([]models.Command, 0, len(c.Commands))
		for _, cmd := range c.Commands {
			result = append(result, cmd)
		}
		return result
	}

	// Cluster has commands: resolve each key against global definitions
	result := make([]models.Command, 0, len(cluster.Commands))
	for key, clusterCmd := range cluster.Commands {
		resolved := clusterCmd

		// If the key exists globally, inherit missing fields
		if globalCmd, ok := c.Commands[key]; ok {
			if resolved.Name == "" {
				resolved.Name = globalCmd.Name
			}
			if resolved.Cmd == "" {
				resolved.Cmd = globalCmd.Cmd
			}
			if len(resolved.MatchServices) == 0 {
				resolved.MatchServices = globalCmd.MatchServices
			}
		}

		// Use the key as fallback name if still empty
		if resolved.Name == "" {
			resolved.Name = key
		}

		result = append(result, resolved)
	}
	return result
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
