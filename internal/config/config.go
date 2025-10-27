package config

import (
	"github.com/mendes11/swarm-browser/internal/core/models"
)

type Config struct {
	ClusterFilePath string
	InitialCluster  string
	Clusters        map[string]models.Cluster
}

var defaultConfig = &Config{
	ClusterFilePath: "./clusters.yml",
	Clusters:        nil,
}

func LoadConfig() Config {
	conf := defaultConfig
	// TODO: Load configs from variables or through Flags
	clusters, err := LoadClustersConfig(conf.ClusterFilePath)
	if err != nil {
		panic(err)
	}
	conf.Clusters = clusters.Clusters
	for k := range conf.Clusters {
		conf.InitialCluster = k
		break
	}
	return *conf
}
