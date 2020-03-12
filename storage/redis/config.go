package redis

import (
	"github.com/fezho/oidc-auth-service/storage"
	"github.com/gomodule/redigo/redis"
	"time"
)

// Redis config for connecting to redis server.
type Config struct {
	// The network type, either tcp or unix.
	// Default is tcp.
	Network string
	// The host:port address of redis server.
	Address string
	// Optional password. Must match the password specified in the
	// require pass server configuration option.
	Password string
	// Database to be selected after connecting to the server.
	DB        int
	KeyPrefix string
	// Maximum number of idle connections in the pool.
	MaxIdle int
	// Close connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout time.Duration

	// TODO: tls.Config
}

func (r *Config) Open(config *storage.Config) (*storage.Storage, error) {
	pool := &redis.Pool{
		MaxIdle:     r.MaxIdle,
		IdleTimeout: r.IdleTimeout,
		Dial: func() (redis.Conn, error) {
			return dialWithDB(r.Network, r.Address, r.Password, r.DB)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	redisConn := &redisConn{
		Pool:       pool,
		keyPrefix:  r.KeyPrefix,
		serializer: config.Serializer,
	}

	return storage.New(redisConn, config), nil
}
