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
	// SweepFrequency is the frequency of task for sweeping expired sessions
	// Default value is 30 seconds
	SweepFrequency time.Duration
}

func (c *Config) Open(config *storage.Config) (*storage.Storage, error) {
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
		serializer:    config.Serializer,
		cancel:        cancel,
		maxAge:        time.Second * time.Duration(config.MaxAge),
	}
	conn.startExpireTask(ctx, c.SweepFrequency)

	return storage.New(conn, config), nil
}
