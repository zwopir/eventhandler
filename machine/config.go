package machine

// CoordinatorConfig represents the settings that specified the executed command
type CoordinatorConfig struct {
	Cmd           string   `yaml:"cmd"`
	CmdArgs       []string `yaml:"cmdargs"`
	Timeout       string   `yaml:"timeout"`
	StdinTemplate string   `yaml:"stdintemplate"`
	Blackout      string   `yaml:"blackout"`
	MaxDispatches int64    `yaml:"maxdispatches"`
}
