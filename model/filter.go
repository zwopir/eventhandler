package model

import (
	"regexp"
	"eventhandler/config"
	"encoding/json"
	"fmt"
	"errors"
)

var (
	RetrieverMissingFieldError error = errors.New("failed to retrieve value from interface")
)


type Filterer interface{
	Match(interface{}) (bool, error)
}

type FilterBattery []Filterer

func newFilters(filters ...Filterer) FilterBattery {
	ret := FilterBattery{}
	for _, f := range filters {
		ret = append(ret, f)
	}
	return ret
}

// Match implements the Filterer interface. It returns a match if all contained Filterer slice elements
// return a match
func (f FilterBattery) Match(v interface{}) (bool, error) {
	for _, f := range f {
		matched, err := f.Match(v)
		if err != nil {
			return false, err
		}
		if ! matched {
			return false, nil
		}
	}
	return true, nil
}

// basicFilter is an unexported basic type that implements the Filterer interface
// its Match method returns the result of the evaluated embedded match function
type basicFilter struct{
	matchFunc func(interface{}) (bool, error)
}

// implement the Filterer interface
func (f basicFilter) Match(v interface{}) (bool, error) {
	return f.matchFunc(v)
}

// newFilter returns a filterer based on the provided func
func newFilter(f func(interface{}) (bool, error)) basicFilter {
	return basicFilter{
		matchFunc: f,
	}
}

// retriever retries a value to be filtered by a Filterer
type retriever interface{
	getValue(v interface{}) ([]byte, error)
}

// envelopeValueRetriever retrieves a value from an envelope struct field
type envelopeValueRetriever struct{
	field string
}

func (r envelopeValueRetriever) getValue(v interface{}) ([]byte, error) {
	e, ok := v.(Envelope)
	if !ok {
		return nil, fmt.Errorf("type assertion of %v to Envelope failed", v)
	}
	switch r.field {
	case "sender":
		return e.Sender, nil
	case "recipient":
		return e.Recipient, nil
	case "payload":
		return e.Payload, nil
	case "signature":
		return e.Signature, nil
	default:
		return nil, RetrieverMissingFieldError
	}
}

func newEnvelopeValueRetriever(field string) envelopeValueRetriever {
	return envelopeValueRetriever{
		field: field,
	}
}

type payloadMessageKeyRetriever struct{
	key string
}

func newPayloadMessageKeyRetriever(key string) payloadMessageKeyRetriever {
	return payloadMessageKeyRetriever{key: key}
}

func (p payloadMessageKeyRetriever) getValue(v interface{}) ([]byte, error) {
	e, ok := v.(Envelope)
	if !ok {
		return nil, fmt.Errorf("type assertion of %v to Envelope failed", v)
	}
	message := Message{}
	err := json.Unmarshal(e.Payload, &message)
	if err != nil {
		return nil, err
	}
	if ret, ok := message[p.key]; ok {
		return []byte(ret), nil
	}
	return nil, RetrieverMissingFieldError
}


func newRegexpFilterer(retriever retriever, regexp *regexp.Regexp) (Filterer, error) {
	filterer := newFilter(
		func(v interface{}) (bool, error) {
			valueToCompare, err := retriever.getValue(v)
			// consider match tries on a non existent field as a non-match
			if err == RetrieverMissingFieldError {
				return false, nil
			}
			// a real error occurred
			if err != nil {
				return false, err
			}
			return regexp.Match(valueToCompare), nil
		},
	)
	return filterer, nil
}

func NewFiltererFromConfig(configFilters []config.Filter) (Filterer, error) {
	var (
		retriever retriever
		matcher   Filterer
		err       error
		re        *regexp.Regexp
	)
	filters := []Filterer{}
	for _, cf := range configFilters {
		field, found := cf.Args["field"]
		if !found {
			return nil, errors.New("mandatory argument 'field' not found in filter configuration.")
		}
		switch cf.Context {
		case "payload":
			retriever = newPayloadMessageKeyRetriever(field)
		case "envelope":
			retriever = newEnvelopeValueRetriever(field)
		default:
			return nil, fmt.Errorf("basicFilter context %s is not supported", cf.Context)
		}
		regexpString, found := cf.Args["regexp"]
		if !found {
			return nil, errors.New("mandatory argument 'regexp' not found in filter configuration")
		}
		switch cf.Type {
		case "regexp":
			re, err = regexp.Compile(regexpString)
			if err != nil {
				return nil, err
			}
			matcher, err = newRegexpFilterer(retriever, re)
		default:
			return nil, fmt.Errorf("basicFilter type %s is not supported", cf.Type)
		}
		filters = append(filters, matcher)
	}
	return newFilters(filters...), nil
}

/*
func newSignatureFilterer() (Filterer, error) {

}
*/


