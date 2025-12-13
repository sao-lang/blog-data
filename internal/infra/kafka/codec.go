package kafka

import "encoding/json"

func EncodeJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func DecodeJSON(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}
