package config

type Config struct {
	ClusterFilePath string
	Clusters        map[string]Cluster
}

var defaultConfig = &Config{
	ClusterFilePath: "./clusters.yaml",
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
	return *conf
}
