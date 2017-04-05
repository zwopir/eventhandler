package model

import (
	"encoding/json"

)

type Message map[string]string

func (m *Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(m)
}
