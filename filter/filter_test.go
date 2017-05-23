package filter

import (
	"eventhandler/model"
	"testing"
)

var (
	modelTT = []struct {
		message       model.Envelope
		expectedMatch bool
		configFilters FilterConfig
	}{
		{
			model.Envelope{
				Sender:    []byte(`a_sender`),
				Recipient: []byte(`a_recipient`),
				Payload:   []byte(`{"check_name":"check_foo"}`),
				Signature: []byte(`sig sig sig`),
			},
			true,
			FilterConfig{
				{
					Context: "payload map",
					Type:    "regexp",
					Args: map[string]string{
						"field":  "check_name",
						"regexp": "check_.+",
					},
				},
			},
		},
		{
			model.Envelope{
				Sender:    []byte(`a_sender`),
				Recipient: []byte(`a_recipient`),
				Payload:   []byte(`{"check_name":"check_foo"}`),
				Signature: []byte(`sig sig sig`),
			},
			false,
			FilterConfig{
				{
					Context: "payload map",
					Type:    "regexp",
					Args: map[string]string{
						"field":  "check_name",
						"regexp": "not_gonna_match_.+",
					},
				},
			},
		},
		{
			model.Envelope{
				Sender:    []byte(`a_sender`),
				Recipient: []byte(`a_recipient`),
				Payload:   []byte(`{"check_name":"check_foo"}`),
				Signature: []byte(`sig sig sig`),
			},
			true,
			FilterConfig{
				{
					Context: "envelope",
					Type:    "regexp",
					Args: map[string]string{
						"field":  "sender",
						"regexp": "a_send.+",
					},
				},
			},
		},
		{
			model.Envelope{
				Sender:    []byte(`a_sender`),
				Recipient: []byte(`a_recipient`),
				Payload:   []byte(`{"check_name":"check_foo"}`),
				Signature: []byte(`sig sig sig`),
			},
			true,
			FilterConfig{
				{
					Context: "payload template",
					Type:    "regexp",
					Args: map[string]string{
						"template": "{{ index . \"check_name\" }}",
						"regexp":   "check_.+",
					},
				},
			},
		},
	}
)

func TestFilters_Match(t *testing.T) {
	for _, tt := range modelTT {
		filters, err := NewFiltererFromConfig(tt.configFilters)
		if err != nil {
			t.Errorf("failed to create basicFilter for %v (%s)",
				tt.configFilters,
				err,
			)
		}
		matched, err := filters.Match(tt.message)
		if err != nil {
			t.Errorf("Match failed with: %s", err)
		}
		if matched != tt.expectedMatch {
			t.Error("expected all filters to match, but got a mismatch")
		}
	}
}

var notCompilingFilter = FilterConfig{
	{
		Context: "payload map",
		Type:    "regexp",
		Args: map[string]string{
			"field":  "check_name",
			"regexp": "((a)",
		},
	},
}

func TestNewFilters(t *testing.T) {
	_, err := NewFiltererFromConfig(notCompilingFilter)
	if err == nil {
		t.Errorf("NewFilters should return an error with regexp %s",
			notCompilingFilter[0].Args["regexp"],
		)
	}
}
