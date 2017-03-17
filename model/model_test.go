package model

import (
	"ehclient/config"
	"testing"
)

var (
	modelTT = []struct {
		message       *Message
		expectedMatch bool
		configFilters []config.Filter
	}{
		{
			&Message{
				"hostname":   "localhost",
				"check_name": "check_whatever",
			},
			true,
			[]config.Filter{
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
		{
			&Message{
				"testkey":         "testvalue",
				"another_testkey": "whatever",
			},
			false,
			[]config.Filter{
				{
					SourceField: "not_provided",
					RegexpMatch: "ignored",
				},
				{
					SourceField: "also_not_provided",
					RegexpMatch: "ignored_too",
				},
			},
		},
		{
			&Message{
				"hostname":   "localhost",
				"check_name": "check_whatever",
			},
			false,
			[]config.Filter{
				{
					SourceField: "hostname",
					RegexpMatch: "whatever",
				},
			},
		},
	}
)

func TestFilters_MatchAll(t *testing.T) {
	for _, tt := range modelTT {
		filters, err := NewFilters(tt.configFilters)
		if err != nil {
			t.Errorf("failed to create filter for %v (%s)",
				tt.configFilters,
				err,
			)
		}
		matched := filters.MatchAll(tt.message)
		if matched != tt.expectedMatch {
			t.Error("expected all filters to match, but got a mismatch")
		}
	}
}

var notCompilingFilter = []config.Filter{
	{
		SourceField: "a_field",
		RegexpMatch: "[[[aa]",
	},
}

func TestNewFilters(t *testing.T) {
	_, err := NewFilters(notCompilingFilter)
	if err == nil {
		t.Errorf("NewFilters should return an error with regexp %s",
			notCompilingFilter[0].RegexpMatch,
		)
	}
}
