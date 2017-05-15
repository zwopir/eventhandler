package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Global  Global  `yaml:"global"`
	Command Command `yaml:"command"`
}

type Global struct {
	Subject     string `yaml:"subject"`
	NatsAddress string `yaml:"natsaddress"`
}

type Command struct {
	Cmd           string   `yaml:"cmd"`
	CmdArgs       []string `yaml:"cmd_args"`
	StdinTemplate string   `yaml:"stdin_template"`
	Filters       []Filter `yaml:"filters"`
	Blackout      string   `yaml:"blackout"`
	MaxDispatches int64    `yaml:"max_dispatches"`
}

type Filter struct {
	Type    string            `yaml:"type"`
	Context string            `yaml:"context"`
	Args    map[string]string `yaml:"args"`
}

func FromFile(configFile string) (*Config, error) {
	var c Config
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
