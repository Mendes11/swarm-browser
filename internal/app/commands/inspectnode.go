package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/core"
	"github.com/mendes11/swarm-browser/internal/core/models"
)

type InspectNodeFailed struct {
	Err error
}

type NodeInspected struct {
	Node models.Node
	Info models.NodeInfo
}

func InspectNode(browser core.SwarmConnector, node models.Node) tea.Cmd {
	return func() tea.Msg {
		nodeInfo, err := browser.InspectNode(node)
		if err != nil {
			return InspectNodeFailed{Err: err}
		}
		return NodeInspected{
			Node: node,
			Info: *nodeInfo,
		}
	}
}
