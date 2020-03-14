package storage

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

// Serializer provides an interface for providing custom serializers for cookie
// values.
type Serializer interface {
	Serialize(src map[interface{}]interface{}) ([]byte, error)
	Deserialize(src []byte, dst map[interface{}]interface{}) error
}

// GobEncoder encodes cookie values using encoding/gob. This is the simplest
// encoder and can handle complex types via gob.Register.
type GobEncoder struct{}

// JSONEncoder encodes cookie values using encoding/json. Users who wish to
// encode complex types need to satisfy the json.Marshaller and
// json.Unmarshaller interfaces.
type JSONEncoder struct{}

// Serialize encodes a value using gob.
func (e GobEncoder) Serialize(src map[interface{}]interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(src); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize decodes a value using gob.
func (e GobEncoder) Deserialize(src []byte, dst map[interface{}]interface{}) error {
	dec := gob.NewDecoder(bytes.NewBuffer(src))
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

// Serialize encodes a value using encoding/json.
func (e JSONEncoder) Serialize(src map[interface{}]interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(src); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize decodes a value using encoding/json.
func (e JSONEncoder) Deserialize(src []byte, dst map[interface{}]interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(src))
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}
