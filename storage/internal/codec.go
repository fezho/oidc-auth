package internal

import (
	"bytes"
	"encoding/gob"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

const defaultHashKey = "Don'tChangeTheKeyValueAnyMoreAfterRunningOnce"

// Encode encodes session values to bytes
func Encode(session *sessions.Session) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(session.Values); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode decodes session values from bytes
func Decode(data []byte, session *sessions.Session) error {
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err := dec.Decode(&session.Values)
	if err != nil {
		return err
	}
	return nil
}

// CodecsFromPairs returns a slice of securecookie.Codec instances,
// if keyPairs is empty, it would use a default hash key to make sure
// at least one Codec is returned.
func CodecsFromPairs(keyPairs []string) []securecookie.Codec {
	if len(keyPairs) == 0 {
		keyPairs = append(keyPairs, defaultHashKey)
	}
	byteKeys := make([][]byte, len(keyPairs))
	for i, keyPair := range keyPairs {
		byteKeys[i] = []byte(keyPair)
	}
	return securecookie.CodecsFromPairs(byteKeys...)
}
