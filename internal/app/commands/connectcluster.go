package commands

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/core"
	"github.com/mendes11/swarm-browser/internal/core/models"
)

type ClusterConnected struct {
	Cluster models.Cluster
	Browser core.ClusterBrowser
	Info    models.NodeInfo
}

type ClusterConnectionFailed struct {
	Err error
}

func ConnectToCluster(cluster models.Cluster) tea.Cmd {
	return func() tea.Msg {
		log.Println("Initializing ClusterBrowser")
		browser := core.New(cluster)
		log.Println("Inspecting Cluster Node")
		nodeInfo, err := browser.InspectNode(cluster.Node)
		if err != nil {
			log.Printf("Failed to inspect cluster node: %v\n", err)
			return ClusterConnectionFailed{
				Err: err,
			}
		}
		log.Println("Successfully Connected to Cluster")
		return ClusterConnected{
			Cluster: cluster,
			Browser: browser,
			Info:    *nodeInfo,
		}
	}
}
