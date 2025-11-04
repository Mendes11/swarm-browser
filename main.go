package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/app"
	"github.com/mendes11/swarm-browser/internal/config"
)

func main() {
	conf := config.LoadConfig()
	app := app.New(conf)
	defer app.Close()

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := tea.NewProgram(app, tea.WithAltScreen()).Run(); err != nil {
		log.Printf("Program exited with error: %v", err)
		panic(err)
	}
}
