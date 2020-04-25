package redis

import (
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/sessions"

	"github.com/fezho/oidc-auth/storage/internal"
)

type redisConn struct {
	Pool      *redis.Pool
	keyPrefix string
}

func dial(address, password string, db int) (redis.Conn, error) {
	var opts []redis.DialOption
	if password != "" {
		opts = append(opts, redis.DialPassword(password))
	}
	if db != 0 {
		opts = append(opts, redis.DialDatabase(db))
	}

	return redis.Dial("tcp", address, opts...)
}

func (c *redisConn) Close() error {
	return c.Pool.Close()
}

func (c *redisConn) getKey(sessionID string) string {
	return c.keyPrefix + sessionID
}

func (c *redisConn) Save(session *sessions.Session) error {
	data, err := internal.Encode(session)
	if err != nil {
		return err
	}

	conn := c.Pool.Get()
	defer conn.Close()

	age := session.Options.MaxAge
	_, err = conn.Do("SETEX", c.getKey(session.ID), age, data)
	return err
}

func (c *redisConn) Load(session *sessions.Session) (bool, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	data, err := conn.Do("GET", c.getKey(session.ID))
	if err != nil {
		return false, err
	}
	if data == nil {
		return false, nil // no data was associated with this key
	}

	b, err := redis.Bytes(data, err)
	if err != nil {
		return false, err
	}

	return true, internal.Decode(b, session)
}

func (c *redisConn) Delete(session *sessions.Session) error {
	conn := c.Pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", c.getKey(session.ID))
	return err
}
