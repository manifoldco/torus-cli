// Package db provides persistent storage and caching of values returned from
// the registry. Sensitive data is stored in the db in its encrypted form.
package db

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"

	"github.com/arigatomachine/cli/identity"

	"github.com/arigatomachine/cli/daemon/envelope"
)

var schemaVersion = []byte{0x01}

// DB is a persistent store for encrypted or non-sensitvie values.
type DB struct {
	db *bolt.DB
}

// NewDB creates a new db or opens an existing db at the given path.
// If the db already exists but has a mismatched version, it will be cleared
// before being returned.
func NewDB(path string) (*DB, error) {
	db := &DB{}

	valid, err := db.initBolt(path)
	if valid && err == nil {
		return db, nil
	}

	if !valid {
		log.Print("DB schema version is incorrect. Clearing db")
	}

	err = os.Remove(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to remove db! Please manually remove %s",
			path)
	}

	valid, err = db.initBolt(path)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, fmt.Errorf("Fatal error initializing db schema version")
	}

	return db, nil
}

// Close closes all db resources
func (db *DB) Close() error {
	return db.db.Close()
}

// initBolt initializes the backing bolt db, checking that the metadata schema
// version matches what we expect
func (db *DB) initBolt(path string) (bool, error) {
	var err error
	db.db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return false, err
	}

	return db.checkMeta()
}

// checkMeta check's the db's metadata, ensuring the version of the db is
// correct, or setting it if it does not exist.
func (db *DB) checkMeta() (bool, error) {
	var version []byte
	err := db.db.Update(func(tx *bolt.Tx) error {
		meta, err := tx.CreateBucketIfNotExists([]byte("meta"))
		if err != nil {
			return err
		}

		version = meta.Get([]byte("version"))
		if version != nil {
			return nil
		}

		err = meta.Put([]byte("version"), schemaVersion)
		if err != nil {
			return err
		}
		version = schemaVersion
		return nil
	})

	return bytes.Equal(version, schemaVersion), err
}

// Set stores the serialized value of env into the db, under key id.
// Stored values are grouped by their type.
func (db *DB) Set(envs ...envelope.Envelope) error {
	return db.db.Update(func(tx *bolt.Tx) error {
		for _, env := range envs {
			id := env.GetID()

			b, err := json.Marshal(env)
			if err != nil {
				return err
			}

			bucket, err := tx.CreateBucketIfNotExists([]byte{id.Type()})
			if err != nil {
				return err
			}
			err = bucket.Put(id[:], b)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// Get returns the value of id in env. It returns an error if id does not exist.
func (db *DB) Get(id *identity.ID, env envelope.Envelope) error {
	return db.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte{id.Type()})
		if bucket == nil {
			return errors.New("ID not found")
		}

		b := bucket.Get(id[:])
		if b == nil {
			return errors.New("ID not found")
		}

		return json.Unmarshal(b, env)
	})
}
