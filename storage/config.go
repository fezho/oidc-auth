package storage

import (
	"fmt"
)

type Type string

const (
	BOLT   Type = "bolt"
	REDIS  Type = "redis"
	MEMORY Type = "memory"
)

type SessionConfig struct {
	// KeyPairs are used to generate securecookie.Codec,
	// Should not change them after application is started,
	// otherwise previously issued cookies will not be able to be decoded.
	// Can be created using securecookie.GenerateRandomKey()
	KeyPairs []string `json:"keyPairs"`
	// Session Max-Age attribute present and given in seconds.
	MaxAge int `json:"maxAge"`
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
