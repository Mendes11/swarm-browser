package models

type Node struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Hostname string `yaml:"hostname" mapstructure:"hostname"`
}
