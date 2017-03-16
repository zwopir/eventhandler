package config

import (
	"testing"
	"reflect"
)


var configTT = []struct{
	fixturesFile string
	expectedConfig *Config
}{
	{"../config_example.yaml",
		&Config{
			Global: Global{
				StanAddress: "localhost:4222",
			},
			Command: Command{
				Cmd:           "cat",
				CmdArgs:       []string{"-"},
				StdinTemplate: "{{ . | printf %v }}",
				Filters: []Filter{
					{
						SourceField: "hostname",
						RegexpMatch: "localhost",
					},
					{
						SourceField: "check_name",
						RegexpMatch: "check_.+",
					},
				},
			},
		},
	},
}

func TestFromFile(t *testing.T) {
	for _, tt := range configTT {
		actual, err := FromFile(tt.fixturesFile)
		if err != nil {
			t.Errorf("can't unmarshal config: %s", err)
		}
		if !reflect.DeepEqual(actual, tt.expectedConfig) {
			t.Errorf("expected %v as config, got %v",
				tt.expectedConfig,
				actual,
			)
		}
	}
}
