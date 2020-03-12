package config

import (
	"encoding/json"
	"fmt"
	"github.com/fezho/oidc-auth-service/storage"
	"github.com/fezho/oidc-auth-service/storage/bolt"
	"github.com/fezho/oidc-auth-service/storage/memory"
	"github.com/fezho/oidc-auth-service/storage/redis"
	"github.com/gorilla/securecookie"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

type Config struct {
	Web     Web     `json:"web"`
	Storage Storage `json:"storage"`
	Logger  Logger  `json:"logger"`
}

func LoadConfigFromFile(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return LoadConfig(data)
}

func LoadConfig(data []byte) (*Config, error) {
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

// Web is the config format for the HTTP server.
type Web struct {
	HTTP           string   `json:"http"`
	HTTPS          string   `json:"https"`
	TLSCert        string   `json:"tlsCert"`
	TLSKey         string   `json:"tlsKey"`
	AllowedOrigins []string `json:"allowedOrigins"`
}

// Storage holds app's storage configuration.
type Storage struct {
	Type    string         `json:"type"`
	Session SessionStorage `json:"session"`
	Config  StorageConfig  `json:"config"`
}

// SessionStorage holds session configuration when being stored
type SessionStorage struct {
	Serializer string   `json:"serializer"`
	KeyPairs   []string `json:"keyPairs"`
	MaxAge     int      `json:"maxAge"`
}

// StorageConfig is a configuration that can create a storage.
type StorageConfig interface {
	Open(config *storage.Config) (*storage.Storage, error)
}

// TODO:  use registration pattern
var storages = map[string]func() StorageConfig{
	"bolt":   func() StorageConfig { return new(bolt.Config) },
	"redis":  func() StorageConfig { return new(redis.Config) },
	"memory": func() StorageConfig { return new(memory.Config) },
}

// UnmarshalJSON allows Storage to implement the unmarshaler interface to
// dynamically determine the type of the storage config.
func (s *Storage) UnmarshalJSON(b []byte) error {
	var store struct {
		Type    string          `json:"type"`
		Session SessionStorage  `json:"session"`
		Config  json.RawMessage `json:"config"`
	}
	if err := json.Unmarshal(b, &store); err != nil {
		return fmt.Errorf("parse storage: %v", err)
	}
	f, ok := storages[store.Type]
	if !ok {
		return fmt.Errorf("unknown storage type %q", store.Type)
	}

	storageConfig := f()
	if len(store.Config) != 0 {
		data := []byte(os.ExpandEnv(string(store.Config)))
		if err := json.Unmarshal(data, storageConfig); err != nil {
			return fmt.Errorf("parse storage config: %v", err)
		}
	}
	*s = Storage{
		Type:   store.Type,
		Config: storageConfig,
	}
	return nil
}

// Logger holds configuration required to customize logging for dex.
type Logger struct {
	// Level sets logging level severity.
	Level string `json:"level"`

	// Format specifies the format to be used for logging.
	Format string `json:"format"`
}

func ToStorageConfig(s SessionStorage) (*storage.Config, error) {
	var serializer securecookie.Serializer
	switch strings.ToLower(s.Serializer) {
	case "gob":
		serializer = securecookie.GobEncoder{}
	case "json":
		serializer = securecookie.JSONEncoder{}
	default:
		return nil, fmt.Errorf("%s serializer is not supported", s.Serializer)
	}

	byteKeyPairs := make([][]byte, len(s.KeyPairs))
	for i, keyPair := range s.KeyPairs {
		byteKeyPairs[i] = []byte(keyPair)
	}

	return &storage.Config{
		Serializer: serializer,
		MaxAge:     s.MaxAge,
		KeyPairs:   byteKeyPairs,
	}, nil
}
