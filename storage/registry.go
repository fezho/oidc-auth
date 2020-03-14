package storage

import (
	"fmt"
	"github.com/gorilla/securecookie"
	"strings"
)

type Type string

const (
	BOLT   Type = "bolt"
	REDIS  Type = "redis"
	MEMORY Type = "memory"
)

type SessionConfig struct {
	SerializerStr string   `json:"serializer"`
	KeyPairStrs   []string `json:"keyPairs"` // TODO: check here for serialize/deserialize [][]byte
	MaxAge        int      `json:"maxAge"`

	Serializer securecookie.Serializer
	// Session Max-Age attribute present and given in seconds.
	// KeyPairs are used to generate securecookie.Codec,
	// Should not change them after application is started,
	// otherwise previously issued cookies will not be able to be decoded.
	// Can be created using securecookie.GenerateRandomKey()
	KeyPairs [][]byte
}

func (c *SessionConfig) Unmarshal() error {
	switch strings.ToLower(c.SerializerStr) {
	case "gob":
		c.Serializer = securecookie.GobEncoder{}
	case "json":
		c.Serializer = securecookie.JSONEncoder{}
	default:
		return fmt.Errorf("%s serializer is not supported", c.Serializer)
	}

	c.KeyPairs = make([][]byte, len(c.KeyPairStrs))
	for i, keyPair := range c.KeyPairStrs {
		c.KeyPairs[i] = []byte(keyPair)
	}

	return nil
}

// Config is a config interface that every Conn SessionConfig would implement it
type Config interface {
	Open() (*Storage, error)
}

type ConfigBuilder func() Config

var configBuilders = map[Type]ConfigBuilder{}

func AddConfigBuilder(t Type, f ConfigBuilder) {
	configBuilders[t] = f
}

func BuildStorageConfig(t Type) (Config, error) {
	f, ok := configBuilders[t]
	if !ok {
		return nil, fmt.Errorf("unknown storage type %q", t)
	}
	return f(), nil
}
