package model

import (
	"eventhandler/config"
	"encoding/json"
	"regexp"
)

type Message map[string]string

func (m *Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(&m)
}

type Filters []FilterFunc

func (f Filters) MatchAll(m *Message) bool {
	for _, f := range f {
		if !f(m) {
			return false
		}
	}
	return true
}

type FilterFunc func(*Message) bool

func newRegexpFilterFunc(regexpString, sourceField string) (FilterFunc, error) {
	re, err := regexp.Compile(regexpString)
	if err != nil {
		return nil, err
	}
	return func(m *Message) bool {
		if value, found := (*m)[sourceField]; found {
			return re.MatchString(value)
		}
		return false
	}, nil
}

func NewFilters(configFilters []config.Filter) (Filters, error) {
	filters := Filters{}
	for _, cf := range configFilters {
		ff, err := newRegexpFilterFunc(cf.RegexpMatch, cf.SourceField)
		if err != nil {
			return nil, err
		}
		filters = append(filters, ff)
	}
	return filters, nil
}
