package devbrowser

import (
	"fmt"
	"os"

	"github.com/mendes11/swarm-browser/internal/core/models"
	"gopkg.in/yaml.v3"
)

// DevConfig represents the complete development configuration
type DevConfig struct {
	Clusters map[string]models.Cluster `yaml:"clusters"`
	Stacks   []StackConfig              `yaml:"stacks"`
}

// StackConfig represents a mock stack with its services
type StackConfig struct {
	Name        string          `yaml:"name"`
	ClusterName string          `yaml:"cluster"`
	Services    []ServiceConfig `yaml:"services"`
}

// ServiceConfig represents a mock service configuration
type ServiceConfig struct {
	ID           string       `yaml:"id,omitempty"`
	Name         string       `yaml:"name"`
	DesiredTasks uint64       `yaml:"desired_tasks"`
	RunningTasks uint64       `yaml:"running_tasks"`
	Tasks        []TaskConfig `yaml:"tasks,omitempty"`
}

// TaskConfig represents a mock task configuration
type TaskConfig struct {
	ID          string `yaml:"id,omitempty"`
	ContainerID string `yaml:"container_id,omitempty"`
	NodeName    string `yaml:"node"`  // Reference to node name in cluster
	Status      string `yaml:"status"`
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*DevConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config DevConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Generate missing service IDs and validate
	for i, stack := range config.Stacks {
		// Validate cluster exists
		if _, exists := config.Clusters[stack.ClusterName]; !exists {
			return nil, fmt.Errorf("stack '%s' references non-existent cluster '%s'", stack.Name, stack.ClusterName)
		}

		for j, service := range stack.Services {
			if service.ID == "" {
				config.Stacks[i].Services[j].ID = fmt.Sprintf("%s-%s-%03d", stack.Name, service.Name, j+1)
			}
			// Ensure running tasks doesn't exceed desired tasks
			if service.RunningTasks > service.DesiredTasks {
				config.Stacks[i].Services[j].RunningTasks = service.DesiredTasks
			}

			// Validate task node references
			cluster := config.Clusters[stack.ClusterName]
			for _, task := range service.Tasks {
				if task.NodeName != "" {
					found := false
					for nodeName := range cluster.Nodes {
						if nodeName == task.NodeName {
							found = true
							break
						}
					}
					if !found {
						return nil, fmt.Errorf("task in service '%s' references non-existent node '%s' in cluster '%s'",
							service.Name, task.NodeName, stack.ClusterName)
					}
				}
			}
		}
	}

	return &config, nil
}

// GetStacksForCluster returns all stacks associated with a specific cluster
func (c *DevConfig) GetStacksForCluster(clusterName string) []StackConfig {
	var stacks []StackConfig
	for _, stack := range c.Stacks {
		if stack.ClusterName == clusterName {
			stacks = append(stacks, stack)
		}
	}
	return stacks
}

// GetNodeFromCluster gets a node by name from a cluster
func (c *DevConfig) GetNodeFromCluster(clusterName, nodeName string) (models.Node, bool) {
	cluster, exists := c.Clusters[clusterName]
	if !exists {
		return models.Node{}, false
	}
	node, exists := cluster.Nodes[nodeName]
	return node, exists
}