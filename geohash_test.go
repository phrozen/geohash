package geohash

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var geohashTests = []struct {
	latitude  float64
	longitude float64
	geohash   string
}{
	{-27.07332578863511, -109.32321101199314, "3e4mbr3q2w39"}, // Chile - Easter Island, Anakena Beach
	{41.90216070037718, 12.453725061736066, "sr2y7kh9bbfk"},   // Italy - Vatican, Saint Peter's Basillica
	{55.753730934309345, 37.61990186254636, "ucfv0j9vp0xz"},   // Moscow - Red Plaza, Lenin's Monument
	{-33.85684190426881, 151.21525191838856, "r3gx2ux9dg0p"},  // Sydney - Opera House
	{19.43265922422016, -99.13317967733457, "9g3w81t7mqpx"},   // Mexico - CDMX Zócalo
}

func TestLocation(t *testing.T) {
	// New
	l := NewLocation(20.643896, -103.416687)
	assert.NotNil(t, l)
	// Latitude, Longitude
	assert.Equal(t, 20.643896, l.Latitude())
	assert.Equal(t, -103.416687, l.Longitude())
}

func TestRegion(t *testing.T) {
	// Setup
	min := NewLocation(-90.0, -180)
	max := NewLocation(90.0, 180.0)
	// New
	r := NewRegion(min, max)
	assert.NotNil(t, r)
	// Center
	c := r.Center()
	assert.Equal(t, 0.0, c.Latitude())
	assert.Equal(t, 0.0, c.Longitude())
	// Min, Max
	assert.Equal(t, min, r.Min())
	assert.Equal(t, max, r.Max())
}

func TestEncode(t *testing.T) {
	for _, v := range geohashTests {
		for i := 0; i < len(v.geohash); i++ {
			assert.Equal(t, v.geohash[0:i+1], Encode(v.latitude, v.longitude, i+1))
		}
	}
}

func TestDecode(t *testing.T) {
	for _, v := range geohashTests {
		for i := 4; i < len(v.geohash); i++ { // Just enough precision to just care about integers
			pos := Decode(v.geohash[0 : i+1])
			assert.Equal(t, math.Round(v.latitude+0.5), math.Round(pos.Center().Latitude()+0.5))
			assert.Equal(t, math.Round(v.longitude+0.5), math.Round(pos.Center().Longitude()+0.5))
		}
	}
}

func TestNeighbours(t *testing.T) {
	borders := "bcdf2368"
	geohash := "999999999999"
	for i := 0; i < len(geohash); i++ {
		neighbours := Neighbours(geohash[0 : i+1])
		for _, v := range neighbours {
			assert.True(t, Valid(v))
			assert.Greater(t, strings.IndexByte(borders, v[len(v)-1]), -1)
		}
	}
}

func TestValid(t *testing.T) {
	for _, v := range geohashTests {
		assert.True(t, Valid(v.geohash))
	}
	invalid := []string{
		"abcdefgh", // contains 'a'
		"ijk12345", // contains 'i'
		"lmn67890", // contains 'l'
		"opqrstuv", // contains 'o'
		"wxyz?!_#", // conains 'special characters'
		"",         // empty string
	}
	for _, v := range invalid {
		assert.False(t, Valid(v))
	}
}
