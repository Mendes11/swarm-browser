package commands

import (
	"fmt"
	"log"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/core/models"
	"github.com/mendes11/swarm-browser/internal/shell"
)

type HookSucceeded struct {
	Cluster models.Cluster
}

type HookFailed struct {
	Cluster models.Cluster
	Err     error
	Output  string
}

func RunClusterHook(cluster models.Cluster, hook models.Hook) tea.Cmd {
	return func() tea.Msg {
		log.Printf("Running hook %q for cluster %s: %s", hook.Name, cluster.Name, hook.Cmd)

		cmd := exec.Command(shell.UserShell(), "-c", hook.Cmd)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Hook %q failed for cluster %s: %v, output: %s", hook.Name, cluster.Name, err, string(output))
			return HookFailed{
				Cluster: cluster,
				Err:     fmt.Errorf("hook %q failed: %w", hook.Name, err),
				Output:  string(output),
			}
		}

		log.Printf("Hook %q succeeded for cluster %s", hook.Name, cluster.Name)
		return HookSucceeded{Cluster: cluster}
	}
}
