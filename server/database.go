package main

import (
	"bytes"

	bolt "go.etcd.io/bbolt"
)

// Database defines a simple Key/Value Store interface
type Database interface {
	Open() error
	Close() error
	Set(key, value string) error
	Get(key string) string
	GetAllByPrefix(prefix string) map[string]string
}

// BoltDB implements Database with a BoltDB backend
type BoltDB struct {
	name string
	bolt *bolt.DB
}

// NewBoltDB creates a new BoltDB handler
func NewBoltDB(name string) *BoltDB {
	return &BoltDB{name: name}
}

// Open creates or opens a new BoltDB database file with the filename: <name>.db
func (db *BoltDB) Open() error {
	// Create/Open database file
	database, err := bolt.Open(db.name+".db", 0600, nil)
	if err != nil {
		return err
	}
	db.bolt = database
	// Create default bucket
	err = db.bolt.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(db.name))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

// Close closes the BoltDB database handler and returns it's error if any
func (db *BoltDB) Close() error {
	return db.bolt.Close()
}

// Set stores data to the given geohash key
func (db *BoltDB) Set(geohash, data string) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.name))
		return b.Put([]byte(geohash), []byte(data))
	})
}

// Get returns the data (if any) stored in the geohash key
func (db *BoltDB) Get(geohash string) string {
	var data bytes.Buffer
	db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.name))
		data.Write(b.Get([]byte(geohash)))
		return nil
	})
	return data.String()
}

// GetAllByPrefix returns all the key/value pairs with the given prefix
func (db *BoltDB) GetAllByPrefix(geohash string) map[string]string {
	region := make(map[string]string)
	db.bolt.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(db.name)).Cursor()
		for k, v := c.Seek([]byte(geohash)); k != nil && bytes.HasPrefix(k, []byte(geohash)); k, v = c.Next() {
			region[string(k)] = string(v)
		}
		return nil
	})
	return region
}
