package models

type Cluster struct {
	Node     `yaml:",inline" mapstructure:",squash"`
	Name     string             `yaml:"name" mapstructure:"name"`
	Nodes    map[string]Node    `yaml:"nodes" mapstructure:"nodes"`
	Commands map[string]Command `yaml:"commands,omitempty" mapstructure:"commands"`
}

func (c *Cluster) GetNodeByHostname(hostname string) (Node, bool) {
	for _, node := range c.Nodes {
		if node.Hostname == hostname {
			return node, true
		}
	}
	return Node{}, false
}

func (c *Cluster) ListNodes() []string {
	names := make([]string, 0, len(c.Nodes))
	for name := range c.Nodes {
		names = append(names, name)
	}
	return names
}
