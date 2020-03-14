package config

import (
	"encoding/json"
	"fmt"
	"github.com/fezho/oidc-auth-service/storage"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
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
	Type   storage.Type   `json:"type"`
	Config storage.Config `json:"config"`
}

// dynamically determine the type of the storage config.
func (s *Storage) UnmarshalJSON(b []byte) error {
	var store struct {
		Type   storage.Type    `json:"type"`
		Config json.RawMessage `json:"config"`
	}
	if err := json.Unmarshal(b, &store); err != nil {
		return fmt.Errorf("parse storage: %v", err)
	}

	cfg, err := storage.BuildStorageConfig(store.Type)
	if err != nil {
		return fmt.Errorf("unknown storage type %q", store.Type)
	}

	if len(store.Config) == 0 {
		return fmt.Errorf("no storage config found")
	}

	data := []byte(os.ExpandEnv(string(store.Config)))
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse storage config: %v", err)
	}

	*s = Storage{
		Type:   store.Type,
		Config: cfg,
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
