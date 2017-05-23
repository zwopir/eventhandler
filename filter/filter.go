package filter

import (
	"bytes"
	"encoding/json"
	"errors"
	"eventhandler/model"
	"eventhandler/verify"
	"fmt"
	"github.com/prometheus/common/log"
	"os"
	"regexp"
	"text/template"
)

var (
	RetrieverMissingFieldError error = errors.New("failed to retrieve value from interface")
)

// FilterConfig represents the filter config
type FilterConfig []FilterSettings

// Filter represents the filter settings, the Args keys and values are specific to the filtering
// implemented in the package "model"
type FilterSettings struct {
	Type    string            `yaml:"type"`
	Context string            `yaml:"context"`
	Args    map[string]string `yaml:"args"`
}

// Filterer
type Filterer interface {
	Match(interface{}) (bool, error)
}

// FilterBattery is a list of Filterer. It implements the Filterer itself
type FilterBattery []Filterer

// newFilterBattery creates a FilterBattery from a list of types that implement Filterer
func newFilterBattery(filters ...Filterer) FilterBattery {
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
		if !matched {
			return false, nil
		}
	}
	return true, nil
}

// basicFilter is an unexported basic type that implements the Filterer interface
// its Match method returns the result of the evaluated embedded match function
type basicFilter struct {
	matchFunc func(interface{}) (bool, error)
}

// implement the Filterer interface
func (f basicFilter) Match(v interface{}) (bool, error) {
	return f.matchFunc(v)
}

// newBasicFilter returns a basicFilter based on the provided func
func newBasicFilter(f func(interface{}) (bool, error)) basicFilter {
	return basicFilter{
		matchFunc: f,
	}
}

// retriever retrieves a value to be filtered by a Filterer
type retriever interface {
	getValue(v interface{}) ([]byte, error)
}

// envelopeValueRetriever retrieves a value from an envelope struct field
type envelopeValueRetriever struct {
	field string
}

// getValue implements the retriever interface
func (r envelopeValueRetriever) getValue(v interface{}) ([]byte, error) {
	e, ok := v.(model.Envelope)
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

// newEnvelopValueRetriever returns a new envelopeValueRetriever
func newEnvelopeValueRetriever(field string) envelopeValueRetriever {
	return envelopeValueRetriever{
		field: field,
	}
}

// payloadMapRetriever retrieves a value from the envelope's payload.
// It is assumed that the payload is marshalable to a map[string]string
type payloadMapRetriever struct {
	key string
}

// newPayloadMapRetriever returns a new payloadMessageKeyRetriever
func newPayloadMapRetriever(key string) payloadMapRetriever {
	return payloadMapRetriever{key: key}
}

// getValue implements the retriever interface
func (p payloadMapRetriever) getValue(v interface{}) ([]byte, error) {
	e, ok := v.(model.Envelope)
	if !ok {
		return nil, fmt.Errorf("type assertion of %v to Envelope failed", v)
	}
	message := map[string]string{}
	err := json.Unmarshal(e.Payload, &message)
	if err != nil {
		return nil, err
	}
	if ret, ok := message[p.key]; ok {
		return []byte(ret), nil
	}
	return nil, RetrieverMissingFieldError
}

// payloadTemplateRetriever retrieves a value from the envelope's payload via template.
// It is assumed that the payload is marshalable from/to json
type payloadTemplateRetriever struct {
	template *template.Template
}

func newPayloadTemplateRetriever(tmplString string) (payloadTemplateRetriever, error) {
	tmpl, err := template.New("PayloadTemplate").Parse(tmplString)
	if err != nil {
		return payloadTemplateRetriever{}, err
	}
	return payloadTemplateRetriever{
		template: tmpl,
	}, nil
}

// getValue implements the retriever interface
func (tr payloadTemplateRetriever) getValue(v interface{}) ([]byte, error) {
	var data interface{}
	e, ok := v.(model.Envelope)
	if !ok {
		return nil, fmt.Errorf("type assertion of %v to Envelope failed", v)
	}
	b := new(bytes.Buffer)
	err := json.Unmarshal(e.Payload, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data: %s", err)
	}
	err = tr.template.Execute(b, data)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// newRegexpFilterer returns a filterer that implements the filterer interface.
// It retrieves the value with the provided retriever and matches it against the provided regexp
func newRegexpFilterer(retriever retriever, regexp *regexp.Regexp) Filterer {
	filterer := newBasicFilter(
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
	return filterer
}

func newSignatureFilterer(verifier *verify.Verifier) Filterer {
	filterer := newBasicFilter(
		func(v interface{}) (bool, error) {
			messageBuffer := new(bytes.Buffer)
			for _, envelopeField := range []string{"sender", "recipient", "payload"} {
				value, err := envelopeValueRetriever{field: envelopeField}.getValue(v)
				if err != nil {
					return false, err
				}
				messageBuffer.Write(value)
			}
			signature, err := envelopeValueRetriever{field: "signature"}.getValue(v)
			if err != nil {
				return false, err
			}
			verifyErr := verifier.Verify(messageBuffer.Bytes(), signature)
			if verifyErr != nil {
				return false, nil
			}
			return true, nil
		},
	)
	return filterer
}

// NewFiltererFromConfig returns a filterBattery, implementing the Filterer interface.
// The basic filterer and retriever are chosen based on the provided filter config
func NewFiltererFromConfig(configFilters FilterConfig) (Filterer, error) {
	var (
		retriever retriever
		matcher   Filterer
		err       error
		re        *regexp.Regexp
	)
	filters := []Filterer{}
	for _, cf := range configFilters {
		switch cf.Context {
		case "payload map":
			field, found := cf.Args["field"]
			if !found {
				return nil, errors.New("mandatory argument 'field' not found in payload filter configuration.")
			}
			retriever = newPayloadMapRetriever(field)
		case "payload template":
			tmplString, found := cf.Args["template"]
			if !found {
				return nil, errors.New("mandatory agument 'template' not found in template filter configuration")
			}
			retriever, err = newPayloadTemplateRetriever(tmplString)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize template value retreiver: %s", err)
			}
		case "envelope":
			field, found := cf.Args["field"]
			if !found {
				return nil, errors.New("mandatory argument 'field' not found in filter configuration.")
			}
			retriever = newEnvelopeValueRetriever(field)
		case "signature":
			// pass
		default:
			log.Fatalf("filter context %q is not implemented", cf.Context)
		}
		switch cf.Type {
		case "regexp":
			regexpString, found := cf.Args["regexp"]
			if !found {
				return nil, errors.New("mandatory argument 'regexp' not found in filter configuration")
			}
			re, err = regexp.Compile(regexpString)
			if err != nil {
				return nil, err
			}
			matcher = newRegexpFilterer(retriever, re)
		case "signature":
			verifyKey, found := cf.Args["verifykey"]
			if !found {
				return nil, errors.New("mandatory argument 'verifykey' not found in filter configuration")
			}
			verifyKeyBuffer, err := os.Open(verifyKey)
			if err != nil {
				return nil, err
			}
			verifier, err := verify.NewVerifier(verifyKeyBuffer)
			if err != nil {
				return nil, err
			}
			matcher = newSignatureFilterer(verifier)
		default:
			log.Fatalf("filter type %q is not implemented", cf.Type)
		}
		filters = append(filters, matcher)
	}
	if len(filters) < 1 {
		return nil, errors.New("filter battery contains no filter")
	}
	return newFilterBattery(filters...), nil
}
