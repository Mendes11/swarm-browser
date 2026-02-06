package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mendes11/swarm-browser/internal/core/models"
)

type ConnectionStatuses string

const (
	Disconnected      = "Disconnected"
	Connecting        = "Connecting"
	Connected         = "Connected"
	RunningHookStatus = "Running Hook"
)

type ClusterInfo struct {
	Cluster  models.Cluster
	NodeInfo models.NodeInfo
	Status   ConnectionStatuses
	Err      error
}

func ClusterInfoView(info ClusterInfo) string {
	labels := []string{
		LabelStyle.Render("Cluster: "),
		LabelStyle.Render("Hostname: "),
		LabelStyle.Render("Status: "),
	}
	values := []string{
		TextStyle.Render(info.Cluster.Name),
		TextStyle.Render(info.Cluster.Host),
		TextStyle.Render(renderConnectionStatus(info.Status)),
	}
	return lipgloss.JoinHorizontal(
		0,
		lipgloss.JoinVertical(0, labels...),
		lipgloss.JoinVertical(0, values...),
	)
}

func renderConnectionStatus(status ConnectionStatuses) string {
	switch status {
	case Disconnected:
		return DisconnectedStyle.Render(string(status))
	case Connecting, RunningHookStatus:
		return ConnectingStyle.Render(string(status))
	case Connected:
		return ConnectedStyle.Render(string(status))
	default:
		return DisconnectedStyle.Render(string(status))
	}
}
