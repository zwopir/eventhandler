package model

import (
	"eventhandler/config"
	"testing"
)

var (
	modelTT = []struct {
		message       Envelope
		expectedMatch bool
		configFilters []config.Filter
	}{
		{
			Envelope{
				Sender:    []byte(`a_sender`),
				Recipient: []byte(`a_recipient`),
				Payload:   []byte(`{"check_name":"check_foo"}`),
				Signature: []byte(`sig sig sig`),
			},
			true,
			[]config.Filter{
				{
					Context: "payload",
					Type:    "regexp",
					Args: map[string]string{
						"field":  "check_name",
						"regexp": "check_.+",
					},
				},
			},
		},
		{
			Envelope{
				Sender:    []byte(`a_sender`),
				Recipient: []byte(`a_recipient`),
				Payload:   []byte(`{"check_name":"check_foo"}`),
				Signature: []byte(`sig sig sig`),
			},
			false,
			[]config.Filter{
				{
					Context: "payload",
					Type:    "regexp",
					Args: map[string]string{
						"field":  "check_name",
						"regexp": "not_gonna_match_.+",
					},
				},
			},
		},
		{
			Envelope{
				Sender:    []byte(`a_sender`),
				Recipient: []byte(`a_recipient`),
				Payload:   []byte(`{"check_name":"check_foo"}`),
				Signature: []byte(`sig sig sig`),
			},
			true,
			[]config.Filter{
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
			t.Errorf("Match failed with %s", err)
		}
		if matched != tt.expectedMatch {
			t.Error("expected all filters to match, but got a mismatch")
		}
	}
}

var notCompilingFilter = []config.Filter{
	{
		Context: "payload",
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
