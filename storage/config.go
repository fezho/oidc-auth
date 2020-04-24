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

// Config is a config interface that every Conn SessionConfig would implement it
type Config interface {
	Open() (*Storage, error)
	SetSecureCookie(val bool)
}

type SessionConfig struct {
	// KeyPairs are used to generate securecookie.Codec,
	// Should not change them after application is started,
	// otherwise previously issued cookies will not be able to be decoded.
	// Can be created using securecookie.GenerateRandomKey()
	KeyPairs []string `json:"keyPairs"`
	// Session Max-Age attribute present and given in seconds.
	MaxAge int `json:"maxAge"`

	// secureCookie indicates if the cookie should be set with the Secure flag, meaning it should
	// only ever be sent over HTTPS. This value is inferred by the scheme of the RedirectURL.
	secureCookie bool
}

func (sc *SessionConfig) SetSecureCookie(val bool) {
	sc.secureCookie = val
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
