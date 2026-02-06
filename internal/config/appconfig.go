package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// AppConfig represents the application-level configuration file
// stored at ~/.config/swarm-browser/config.yml.
type AppConfig struct {
	ClustersFile string `yaml:"clusters_file" mapstructure:"clusters_file"`
	ClustersURL  string `yaml:"clusters_url" mapstructure:"clusters_url"`
}

// LoadAppConfig reads the app config from ~/.config/swarm-browser/config.yml.
// Returns a zero-value AppConfig if the file does not exist.
func LoadAppConfig() (AppConfig, error) {
	dir, err := ConfigDir()
	if err != nil {
		return AppConfig{}, fmt.Errorf("failed to get config dir: %w", err)
	}
	return loadAppConfigFrom(dir)
}

func loadAppConfigFrom(dir string) (AppConfig, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yml")
	v.AddConfigPath(dir)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return AppConfig{}, nil
		}
		return AppConfig{}, fmt.Errorf("failed to read app config: %w", err)
	}

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return AppConfig{}, fmt.Errorf("failed to parse app config: %w", err)
	}
	return cfg, nil
}
