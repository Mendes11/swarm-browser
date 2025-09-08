package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Node struct {
	Host     string `yaml:"host"`
	Hostname string `yaml:"hostname"`
}

type Cluster struct {
	Name  string          `yaml:"name"`
	Host  string          `yaml:"host"`
	Nodes map[string]Node `yaml:"nodes"`
}

type ClustersConfig struct {
	Clusters map[string]Cluster `yaml:"clusters"`
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

func (c *ClustersConfig) GetCluster(name string) (*Cluster, bool) {
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

func (c *Cluster) GetNode(name string) (*Node, bool) {
	node, exists := c.Nodes[name]
	return &node, exists
}

func (c *Cluster) GetNodeByHostname(hostname string) (*Node, bool) {
	for _, node := range c.Nodes {
		if node.Hostname == hostname {
			return &node, true
		}
	}
	return nil, false
}

func (c *Cluster) ListNodes() []string {
	names := make([]string, 0, len(c.Nodes))
	for name := range c.Nodes {
		names = append(names, name)
	}
	return names
}