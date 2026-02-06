/*
Copyright © 2025 Rafael Mendes P. Bachiega rafaelmpb11@hotmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mendes11/swarm-browser/internal/app"
	"github.com/mendes11/swarm-browser/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var clustersFile string

var rootCmd = &cobra.Command{
	Use:   "swarm-browser",
	Short: "Terminal UI for browsing Docker Swarm clusters",
	Long: `Swarm Browser is a Terminal User Interface (TUI) application for browsing
and managing Docker Swarm clusters. It provides an interactive way to navigate
through Docker Swarm stacks, services, and tasks, with the ability to attach
directly to containers.`,
	RunE: runApp,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// SetVersionInfo sets version information displayed by --version.
func SetVersionInfo(version, commit, date, builtBy string) {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built at: %s, built by: %s)", version, commit, date, builtBy)
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&clustersFile, "clusters", "",
		"path to clusters.yml file (env: SWARM_BROWSER_CLUSTERS)",
	)
}

func runApp(cmd *cobra.Command, args []string) error {
	v, err := loadClustersViper()
	if err != nil {
		return fmt.Errorf("failed to load clusters config: %w", err)
	}

	clustersConfig, err := config.LoadClustersConfigFromViper(v)
	if err != nil {
		return fmt.Errorf("failed to parse clusters config: %w", err)
	}

	conf := config.Config{
		Clusters: clustersConfig.Clusters,
		Commands: clustersConfig.Commands,
		Hooks:    clustersConfig.Hooks,
	}

	// Load persisted state to restore last-used cluster
	state, err := config.LoadState()
	if err != nil {
		log.Printf("Warning: failed to load state: %v", err)
	}
	if state.LastCluster != "" {
		if _, exists := conf.Clusters[state.LastCluster]; exists {
			conf.InitialCluster = state.LastCluster
		}
	}
	// Fallback: pick the first cluster from the map
	if conf.InitialCluster == "" {
		for k := range conf.Clusters {
			conf.InitialCluster = k
			break
		}
	}

	application := app.New(conf)
	defer application.Close()

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer f.Close()

	if _, err := tea.NewProgram(application, tea.WithAltScreen()).Run(); err != nil {
		log.Printf("Program exited with error: %v", err)
		return err
	}
	return nil
}

// loadClustersViper creates and configures a Viper instance for the clusters config file.
//
// Precedence (highest to lowest):
//  1. --clusters flag
//  2. SWARM_BROWSER_CLUSTERS environment variable
//  3. ./clusters.yml (current working directory)
//  4. UserConfigDir/swarm-browser/clusters.yml
func loadClustersViper() (*viper.Viper, error) {
	v := viper.New()

	explicitPath := clustersFile
	if explicitPath == "" {
		explicitPath = os.Getenv("SWARM_BROWSER_CLUSTERS")
	}

	if explicitPath != "" {
		v.SetConfigFile(explicitPath)
	} else {
		v.SetConfigName("clusters")
		v.SetConfigType("yml")

		v.AddConfigPath(".")

		if userConfigDir, err := os.UserConfigDir(); err == nil {
			v.AddConfigPath(fmt.Sprintf("%s/swarm-browser", userConfigDir))
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("could not read clusters config: %w", err)
	}

	return v, nil
}
