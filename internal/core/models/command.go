package models

// Command represents a predefined command available in the command picker.
type Command struct {
	Name          string   `yaml:"name" mapstructure:"name"`
	Cmd           string   `yaml:"cmd" mapstructure:"cmd"`
	MatchServices []string `yaml:"match_services,omitempty" mapstructure:"match_services"`
}
