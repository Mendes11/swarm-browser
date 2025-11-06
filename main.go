package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/app"
	"github.com/mendes11/swarm-browser/internal/config"
)

// Version variables - set by goreleaser at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// Define command line flags
	versionFlag := flag.Bool("version", false, "Print version information")
	versionShortFlag := flag.Bool("v", false, "Print version information")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Swarm Browser - Terminal UI for Docker Swarm\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  swarm-browser [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nConfiguration:\n")
		fmt.Fprintf(os.Stderr, "  Swarm Browser looks for a 'clusters.yml' file in the current directory\n")
		fmt.Fprintf(os.Stderr, "  to configure cluster connections.\n\n")
		fmt.Fprintf(os.Stderr, "For more information, visit: https://github.com/Mendes11/swarm-browser\n")
	}

	// Parse flags
	flag.Parse()

	// Handle version flag
	if *versionFlag || *versionShortFlag {
		fmt.Printf("swarm-browser version %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built at: %s\n", date)
		fmt.Printf("  built by: %s\n", builtBy)
		os.Exit(0)
	}

	// Normal application startup
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
