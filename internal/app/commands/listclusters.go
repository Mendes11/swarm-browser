package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/core/models"
)

// ClusterTableRow represents a cluster in the table display
type ClusterTableRow struct {
	Name      string // Cluster identifier
	Host      string // Primary manager host
	NodeCount int    // Number of nodes
	IsCurrent bool   // Currently connected
}

// ClustersListed is sent when the cluster list is ready for display
type ClustersListed struct {
	Clusters       []ClusterTableRow
	CurrentCluster string
}

// ListClusters returns a command that prepares cluster data from config for display
func ListClusters(clusters map[string]models.Cluster, currentClusterName string) tea.Cmd {
	return func() tea.Msg {
		var clusterRows []ClusterTableRow

		for name, cluster := range clusters {
			row := ClusterTableRow{
				Name:      name,
				Host:      cluster.Host,
				NodeCount: len(cluster.Nodes),
				IsCurrent: name == currentClusterName,
			}
			clusterRows = append(clusterRows, row)
		}

		return ClustersListed{
			Clusters:       clusterRows,
			CurrentCluster: currentClusterName,
		}
	}
}