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
var clustersURL string

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
	rootCmd.PersistentFlags().StringVar(
		&clustersURL, "clusters-url", "",
		"URL to fetch clusters config from (cached locally on success)",
	)
}

func runApp(cmd *cobra.Command, args []string) error {
	appCfg, err := config.LoadAppConfig()
	if err != nil {
		log.Printf("Warning: failed to load app config: %v", err)
	}

	v, err := loadClustersViper(appCfg)
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
//  1. --clusters-url flag
//  2. --clusters flag
//  3. clusters_url from config.yml
//  4. SWARM_BROWSER_CLUSTERS environment variable
//  5. clusters_file from config.yml
//  6. ./clusters.yml (current working directory)
//  7. UserConfigDir/swarm-browser/clusters.yml
func loadClustersViper(appCfg config.AppConfig) (*viper.Viper, error) {
	v := viper.New()

	// 1. --clusters-url flag (highest precedence)
	if clustersURL != "" {
		return loadClustersFromURL(v, clustersURL)
	}

	// 2. --clusters flag
	if clustersFile != "" {
		v.SetConfigFile(clustersFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("could not read clusters config: %w", err)
		}
		return v, nil
	}

	// 3. clusters_url from config.yml
	if appCfg.ClustersURL != "" {
		return loadClustersFromURL(v, appCfg.ClustersURL)
	}

	// 4. SWARM_BROWSER_CLUSTERS env var
	if envPath := os.Getenv("SWARM_BROWSER_CLUSTERS"); envPath != "" {
		v.SetConfigFile(envPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("could not read clusters config: %w", err)
		}
		return v, nil
	}

	// 5. clusters_file from config.yml
	if appCfg.ClustersFile != "" {
		v.SetConfigFile(appCfg.ClustersFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("could not read clusters config: %w", err)
		}
		return v, nil
	}

	// 6-7. Default search paths
	v.SetConfigName("clusters")
	v.SetConfigType("yml")
	v.AddConfigPath(".")
	if configDir, err := config.ConfigDir(); err == nil {
		v.AddConfigPath(configDir)
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("could not read clusters config: %w", err)
	}
	return v, nil
}

func loadClustersFromURL(v *viper.Viper, url string) (*viper.Viper, error) {
	cachePath, err := config.FetchClustersFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to load clusters from URL: %w", err)
	}
	v.SetConfigFile(cachePath)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("could not read fetched clusters config: %w", err)
	}
	return v, nil
}
