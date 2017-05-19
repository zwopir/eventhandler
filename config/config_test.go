package config

import (
	"reflect"
	"testing"
)

var configTT = []struct {
	fixturesFile   string
	expectedConfig *Config
}{
	{"../config_example.yaml",
		&Config{
			Global: Global{
				NatsAddress: "nats://127.0.0.1:4222",
				Subject:     "eventhandler",
			},
			Command: Command{
				Cmd:           "/bin/cat",
				CmdArgs:       []string{"-"},
				StdinTemplate: "{{ . | printf \"%v\" }}",
				Filters: []Filter{
					{
						Context: "payload",
						Type:    "regexp",
						Args: map[string]string{
							"field":  "check_name",
							"regexp": "check_.+",
						},
					},
					{
						Context: "envelope",
						Type:    "regexp",
						Args: map[string]string{
							"field":  "sender",
							"regexp": "nagios.example.com",
						},
					},
					{
						Context: "envelope",
						Type:    "regexp",
						Args: map[string]string{
							"field":  "recipient",
							"regexp": "me.example.com",
						},
					},
					{
						Context: "signature",
						Type:    "signature",
						Args: map[string]string{
							"verifykey": "verify/testdata/public.key",
						},
					},
				},
				Blackout:      "5s",
				MaxDispatches: 3,
				Timeout:       "2s",
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
