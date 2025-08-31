package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/app"
	"github.com/moby/moby/client"
)

func main() {
	// The app will first have to setup an SSH tunnel to the remote manager's socket.
	cli, err := client.NewClientWithOpts(client.WithHost("unix:///tmp/remote-docker.sock"), client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	app := app.NewApp(cli)
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := tea.NewProgram(app, tea.WithAltScreen()).Run(); err != nil {
		panic(err)
	}
}
