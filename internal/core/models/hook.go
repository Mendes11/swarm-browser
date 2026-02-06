package models

type Hook struct {
	Name string `yaml:"name" mapstructure:"name"`
	Cmd  string `yaml:"cmd" mapstructure:"cmd"`
}
