package server

import (
	"encoding/json"
	"io"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// This is a simple GRPC JSON Protobuf marshaller that just uses standard encoding/json for everything
type JSONMarshaler struct{}

func (jm *JSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jm *JSONMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (jm *JSONMarshaler) NewDecoder(r io.Reader) gwruntime.Decoder {
	return json.NewDecoder(r)
}

func (jm *JSONMarshaler) NewEncoder(w io.Writer) gwruntime.Encoder {
	return json.NewEncoder(w)
}

func (jm *JSONMarshaler) ContentType() string {
	return "application/json"
}
