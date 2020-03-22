package bolt

import (
	"context"
	"github.com/boltdb/bolt"
	"github.com/fezho/oidc-auth-service/storage/internal"
	"github.com/gorilla/sessions"
	"time"
)

type boltConn struct {
	db *bolt.DB

	bucketName    []byte
	ttlBucketName []byte

	maxAge time.Duration

	// This is called once the Close method is called to signal goroutines
	cancel context.CancelFunc
}

func (c *boltConn) Close() error {
	c.cancel()
	return c.db.Close()
}

func (c *boltConn) Save(session *sessions.Session) error {
	data, err := internal.Encode(session)
	if err != nil {
		return err
	}

	err = c.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(c.bucketName).Put([]byte(session.ID), data); err != nil {
			return err
		}

		return tx.Bucket(c.ttlBucketName).Put([]byte(time.Now().UTC().Format(time.RFC3339Nano)), []byte(session.ID))
	})

	return err
}

func (c *boltConn) Load(session *sessions.Session) (exists bool, err error) {
	err = c.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(c.bucketName).Get([]byte(session.ID))
		if data == nil {
			return nil
		}

		err := internal.Decode(data, session)
		if err != nil {
			return err
		}

		exists = true
		return nil
	})
	return
}

func (c *boltConn) Delete(session *sessions.Session) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(c.bucketName).Delete([]byte(session.ID))
	})
}
