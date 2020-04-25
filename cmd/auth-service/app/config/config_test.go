package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/kylelemons/godebug/pretty"

	"github.com/fezho/oidc-auth/cmd/auth-service/app/config"
	"github.com/fezho/oidc-auth/storage"
	"github.com/fezho/oidc-auth/storage/bolt"
	"github.com/fezho/oidc-auth/storage/redis"
)

func TestLoadConfig(t *testing.T) {
	rawConfig := []byte(`
web:
  http: 127.0.0.1:8080
storage:
  type: bolt
  config:
    path: "/tmp/data.bin"
    bucketName: "session"
    maxAge: 1000
    keyPairs: 
      - "key1"
      - "key2"
logger:
  level: "debug"
  format: "json"
`)

	want := config.Config{
		Web: config.Web{
			HTTP: "127.0.0.1:8080",
		},
		Storage: config.Storage{
			Type: "bolt",
			Config: &bolt.Config{
				Path:       "/tmp/data.bin",
				BucketName: "session",
				SessionConfig: storage.SessionConfig{
					KeyPairs: []string{"key1", "key2"},
					MaxAge:   1000,
				},
			},
		},
		Logger: config.Logger{
			Level:  "debug",
			Format: "json",
		},
	}

	c, err := config.LoadConfig(rawConfig)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if diff := pretty.Compare(c, want); diff != "" {
		t.Errorf("got!=want: %s", diff)
	}
}

func TestLoadConfigWithEnv(t *testing.T) {
	redisPasswordEnv := os.Getenv("REDIS_PASSWORD")
	if redisPasswordEnv == "" {
		t.Skipf("test environment variable %q not set, skipping", "REDIS_PASSWORD")
	}
	rawConfig := []byte(`
web:
  http: 127.0.0.1:8080
storage:
  type: redis
  config:
    address: 127.0.0.1:6379
    password: "${REDIS_PASSWORD}"
    db: 3
    keyPrefix: "session-"
    maxAge: 1000
logger:
  level: "debug"
  format: "json"
`)

	want := config.Config{
		Web: config.Web{
			HTTP: "127.0.0.1:8080",
		},
		Storage: config.Storage{
			Type: "redis",
			Config: &redis.Config{
				Address:   "127.0.0.1:6379",
				DB:        3,
				Password:  redisPasswordEnv,
				KeyPrefix: "session-",
				SessionConfig: storage.SessionConfig{
					MaxAge: 1000,
				},
			},
		},
		Logger: config.Logger{
			Level:  "debug",
			Format: "json",
		},
	}

	c, err := config.LoadConfig(rawConfig)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if diff := pretty.Compare(c, want); diff != "" {
		t.Errorf("got!=want: %s", diff)
	}
}

func TestValidConfiguration(t *testing.T) {
	cfg := config.Config{
		Web: config.Web{
			HTTP: "localhost:8000",
		},
		OIDC: config.OIDC{
			Issuer:        "dex.io/dex",
			RedirectURL:   "auth-service:8080/callback",
			ClientID:      "my-app",
			ClientSecret:  "my-secret",
			UsernameClaim: "email",
		},
		Storage: config.Storage{
			Type: "bolt",
			Config: &bolt.Config{
				Path:       "/tmp/data.bin",
				BucketName: "session",
				SessionConfig: storage.SessionConfig{
					KeyPairs: []string{"key1", "key2"},
					MaxAge:   1000,
				},
			},
		},
		Logger: config.Logger{
			Level:  "info",
			Format: "json",
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("this configuration should have been valid: %v", err)
	}
}

func TestInValidConfiguration(t *testing.T) {
	cfg := config.Config{}
	err := cfg.Validate()
	if err == nil {
		t.Fatalf("this configuration should have been invalid: %v", err)
	}

	got := err.Error()
	wanted := `invalid Config:
	-	no openID connect issuer specified
	-	no openID connect redirect url specified, or trailing slash is not allowed
	-	no openID connect client id specified
	-	no openID connect client secret specified
	-	no openID connect user name claim specified
	-	no storage supplied in config file
	-	must supply a HTTP/HTTPS  address to listen on`
	if got != wanted {
		fmt.Println(got)
		t.Fatalf("Expected error message to be %q, got %q", wanted, got)
	}
}
