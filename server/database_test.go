package main

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseSuccess(t *testing.T) {
	// Test Open
	db := NewBoltDB("testing")
	assert.NoError(t, db.Open())
	// Remove database file afterwards
	defer func() {
		os.Remove("testing.db")
	}()
	// Test Set Values
	for i := 0; i < 100; i++ {
		assert.NoError(t, db.Set(strconv.Itoa(i), strconv.Itoa(i)))
	}
	// Test Get Values
	for i := 0; i < 100; i++ {
		assert.Equal(t, strconv.Itoa(i), db.Get(strconv.Itoa(i)))
	}
	// Test GetAllByPrefix
	data := db.GetAllByPrefix("9")
	for k, v := range data {
		assert.Equal(t, "9", k[0:1])
		assert.Equal(t, k, v)
	}
	// Test Close
	assert.NoError(t, db.Close())
}

func TestDatabaseFailure(t *testing.T) {
	db := NewBoltDB("!/._.#?")
	assert.Error(t, db.Open(), "Should fail because invalid characters")
}
