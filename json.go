package http_go

import (
	"encoding/json"
)

// MarshalJSON marshals a value to JSON bytes
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// UnmarshalJSON unmarshals JSON bytes to a value
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
