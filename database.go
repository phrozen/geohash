package geohash

import (
	"bytes"
	"errors"
	"fmt"
)

// DB ...
type DB struct {
	gh   *GeoHash
	root *node
}

// NewDatabase ...
func NewDatabase(gh *GeoHash) *DB {
	if gh == nil {
		gh = globe
	}
	return &DB{gh: gh, root: &node{}}
}

//
type node struct {
	geohash string
	parent  *node
	child   [32]*node
	data    map[string]string
}

func (n *node) String() string {
	var children bytes.Buffer
	for i := range n.child {
		if n.child[i] != nil {
			children.WriteByte(base32[i])
		}
	}
	return fmt.Sprintf("Geohash: %s\nChildren: %s\nData:%v", n.geohash, children.String(), n.data)
}

func (db *DB) walk(geohash string, n *node, create bool) *node {
	if len(geohash) == 0 {
		return n
	}
	i := bytes.IndexByte(base32, geohash[0])
	/* Not needed if previous validation
	if i == -1 {
		return nil
	}
	*/
	if n.child[i] == nil {
		if !create {
			return nil // Not found
		}
		n.child[i] = &node{geohash: n.geohash + string(geohash[0]), parent: n, data: make(map[string]string)}
	}
	return db.walk(geohash[1:], n.child[i], create)
}

func (db *DB) retrieveChildren(n *node, results *[]string) {
	if n == nil {
		return
	}
	if len(n.data) > 0 {
		*results = append(*results, n.geohash)
	}
	for _, c := range n.child {
		if c != nil {
			db.retrieveChildren(c, results)
		}
	}
}

// Set ...
func (db *DB) Set(geohash string, key string, value string) error {
	if err := Validate(geohash); err != nil {
		return err
	}
	n := db.walk(geohash, db.root, true)
	if n == nil {
		return errors.New("Invalid character in geohash (base32)")
	}
	n.data[key] = value
	return nil
}

// Get ...
func (db *DB) Get(geohash string, key string) (string, error) {
	if err := Validate(geohash); err != nil {
		return "", err
	}
	n := db.walk(geohash, db.root, false)
	if n == nil {
		return "", errors.New("Geohash not found")
	}
	if value, exist := n.data[key]; exist {
		return value, nil
	}
	return "", errors.New("Key not found at geohash")
}

// GetAllData ...
func (db *DB) GetAllData(geohash string) (map[string]string, error) {
	if err := Validate(geohash); err != nil {
		return nil, err
	}
	n := db.walk(geohash, db.root, false)
	if n == nil {
		return nil, errors.New("Geohash not found")
	}
	return n.data, nil
}

// GetAllChildren ...
func (db *DB) GetAllChildren(geohash string) ([]string, error) {
	if err := Validate(geohash); err != nil {
		return nil, err
	}
	n := db.walk(geohash, db.root, false)
	if n == nil {
		return nil, errors.New("Geohash not found")
	}
	results := make([]string, 0)
	db.retrieveChildren(n, &results)
	return results, nil
}

// Delete ...
func (db *DB) Delete(geohash string, key string) error {
	if err := Validate(geohash); err != nil {
		return err
	}
	n := db.walk(geohash, db.root, false)
	if n == nil {
		return errors.New("Geohash not found")
	}
	if _, exist := n.data[key]; exist {
		delete(n.data, key)
		return nil
	}
	return errors.New("Key not found at geohash")
}

// Clear all data at a given geohash
func (db *DB) Clear(geohash string) error {
	if err := Validate(geohash); err != nil {
		return err
	}
	n := db.walk(geohash, db.root, false)
	if n == nil {
		return errors.New("Geohash not found")
	}
	n.data = make(map[string]string)
	return nil
}
