package config

import (
	"github.com/mendes11/swarm-browser/internal/core/models"
	"github.com/spf13/viper"
)

type Config struct {
	InitialCluster string
	Clusters       map[string]models.Cluster
	Commands       map[string]models.Command
	Hooks          map[string]models.Hook
}

// LoadClustersConfigFromViper loads the clusters configuration from the file
// that Viper resolved. It uses direct YAML parsing instead of Viper's
// Unmarshal to avoid losing empty map entries (e.g. "postgres-console: {}").
func LoadClustersConfigFromViper(v *viper.Viper) (*ClustersConfig, error) {
	return LoadClustersConfig(v.ConfigFileUsed())
}
