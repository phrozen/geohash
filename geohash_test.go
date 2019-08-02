package geohash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func LocationTest(t *testing.T) {
	l := NewLocation(20.643896, -103.416687)
	assert.NotNil(t, l)
	assert.Equal(t, l.Latitude(), 20.643896)
	assert.Equal(t, l.Longitude(), -103.416687)
}

func RegionTest(t *testing.T) {
	r := NewRegion(NewLocation(-90.0, -180), NewLocation(90.0, 180.0))
	assert.NotNil(t, r)
	c := r.Center()
	assert.Equal(t, c.Latitude(), 0.0)
	assert.Equal(t, c.Longitude(), 0.0)
}

func EncodeTest(t *testing.T) {
	assert.Equal(t, Encode(20.644012, -103.416807, 12), "9ewmqwnhjdz8")
}

func DecodeTest(t *testing.T) {
	
}
