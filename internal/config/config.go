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

// LoadClustersConfigFromViper unmarshals the clusters configuration from a Viper instance.
func LoadClustersConfigFromViper(v *viper.Viper) (*ClustersConfig, error) {
	var cfg ClustersConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
