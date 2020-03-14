package bolt

import (
	"context"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fezho/oidc-auth-service/storage"
	"time"
)

type Config struct {
	// Path is the file path where the database file will be stored.
	Path string
	// BucketName represents the name of the bucket which contains sessions.
	BucketName string

	// SweepFrequency is the frequency for running task to sweep expired sessions,
	// if it's zero or less, means do not running sweep task.
	SweepFrequency time.Duration

	// TODO: need point?
	*storage.SessionConfig `json:",inline"`
}

func init() {
	storage.AddConfigBuilder(storage.BOLT, func() storage.Config { return new(Config) })
}

func (c *Config) Open() (*storage.Storage, error) {
	err := c.SessionConfig.Unmarshal()
	if err != nil {
		return nil, err
	}

	db, err := bolt.Open(c.Path, 0666, nil)
	if err != nil {
		return nil, err
	}

	bucket := []byte(c.BucketName)
	ttlBucket := []byte(c.BucketName + "-ttl")

	err = db.Update(func(tx *bolt.Tx) error {
		// create session bucket
		_, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return fmt.Errorf("create bucket %s error: %v", c.BucketName, err)
		}

		// create a ttl bucket to record session ttl
		_, err = tx.CreateBucketIfNotExists(ttlBucket)
		if err != nil {
			return fmt.Errorf("create bucket %s error: %v", string(ttlBucket), err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	conn := &boltConn{
		db:            db,
		bucketName:    bucket,
		ttlBucketName: ttlBucket,
		serializer:    c.Serializer,
		cancel:        cancel,
		maxAge:        time.Second * time.Duration(c.MaxAge),
	}

	if c.SweepFrequency > 0 {
		conn.startSweeping(ctx, c.SweepFrequency)
	}

	return storage.New(conn, c.SessionConfig), nil
}