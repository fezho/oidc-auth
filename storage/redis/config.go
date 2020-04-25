package redis

import (
	"github.com/fezho/oidc-auth/storage"
	"github.com/gomodule/redigo/redis"
	"time"
)

// Redis config for connecting to redis server.
type Config struct {
	storage.SessionConfig `json:",inline"`

	// The host:port address of redis server.
	Address string
	// Optional password. Must match the password specified in the
	// require pass server configuration option.
	Password string
	// Database to be selected after connecting to the server.
	DB int
	// key prefix for storing session
	KeyPrefix string
}

func init() {
	storage.AddConfigBuilder(storage.REDIS, func() storage.Config { return new(Config) })
}

func (c *Config) Open() (*storage.Storage, error) {
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return dial(c.Address, c.Password, c.DB)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	redisConn := &redisConn{
		Pool:      pool,
		keyPrefix: c.KeyPrefix,
	}

	return storage.New(redisConn, c.SessionConfig), nil
}
