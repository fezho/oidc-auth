package bolt

import (
	"bytes"
	"context"
	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"time"
)

func (c *boltConn) startSweeping(ctx context.Context, frequency time.Duration) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(frequency):
				if err := c.sweep(c.maxAge); err != nil {
					log.Errorf("bolt db: sweep expired session task failed: %v", err)
				}
			}
		}
	}()
}

func (c *boltConn) sweep(maxAge time.Duration) error {
	keys, ttlKeys, err := c.getExpired(maxAge)
	if err != nil || len(keys) == 0 {
		return err
	}

	return c.db.Update(func(tx *bolt.Tx) (err error) {
		bucket := tx.Bucket(c.bucketName)
		for _, key := range keys {
			if err = bucket.Delete(key); err != nil {
				return
			}
		}

		ttlBucket := tx.Bucket(c.ttlBucketName)
		for _, key := range ttlKeys {
			if err = ttlBucket.Delete(key); err != nil {
				return err
			}
		}

		return
	})
}

func (c *boltConn) getExpired(maxAge time.Duration) (keys, ttlKeys [][]byte, err error) {
	err = c.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(c.ttlBucketName).Cursor()

		max := []byte(time.Now().UTC().Add(-maxAge).Format(time.RFC3339Nano))
		for k, v := c.First(); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			keys = append(keys, v)
			ttlKeys = append(ttlKeys, k)
		}
		return nil
	})
	return
}
