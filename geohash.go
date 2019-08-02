// Package geohash ...
// 
// 
// The default geohash specification covers the entire globe from [-90, -180] up to [90, 180]
// with a variable precision. geohash can be applied to other coordinate systems, but the
// geohashes will only work within that coordinate system.
//
// From: https://en.wikipedia.org/wiki/Geohash
// geohash length	lat bits	lng bits	lat error	lng error	km error
// 			1			2			3		 ±23		  ±23	 	  ±2500
// 			2			5			5		 ±2.8	 	  ±5.6	 	  ±630
// 			3			7			8		 ±0.70	 	  ±0.70	  	  ±78
// 			4			10			10		 ±0.087	 	  ±0.18	  	  ±20
// 			5			12			13		 ±0.022	 	  ±0.022	  ±2.4
// 			6			15			15		 ±0.0027	  ±0.0055	  ±0.61
// 			7			17			18		 ±0.00068	  ±0.00068	  ±0.076
// 			8			20			20		 ±0.000085	  ±0.00017	  ±0.019package geohash
package geohash

import (
	"bytes"
	"math"
)

const (
	//Base32 is the dictionary of characters for generating hashes
	base32 = []byte("0123456789bcdefghjkmnpqrstuvwxyz")
	// Bitmask positions for 5 bit base32 encoding
	// []byte{ 0b10000, 0b01000, 0b00100, 0b00010, 0b00001 }
	// []byte{ 0x10, 0x08, 0x04, 0x02, 0x01 }
	bits = []byte{16, 8, 4, 2, 1}
)

// Location is a coordinate pair of latitude, longitude (y, x)
type Location struct {
	lat, lon float64
}

// NewLocation creates a new location (point) with the given coordinates
func NewLocation(latitude, longitude float64) Location {
	return Location{lat: latitude, lon: longitude}
}

// Latitude returns the latitude of the given Location
func (loc Location) Latitude() float64 {
	return loc.lat
}

// Longitude returns the longitude of the given Location
func (loc Location) Longitude() float64 {
	return loc.lon
}

// Region is a bounding box representation of a given area
type Region struct {
	min, max Location
}

// NewRegion defines a new region with 'min' being the South-West (bottom-left) corner
// and 'max' being the North-East (top-right) corner of the box.
func NewRegion(min, max Location) Region {
	return Region{min: min, max: max}
}

// Min returns the South-West location of the Region (bottom-left)
func (r Region) Min() Location {
	return r.min
}

// Max returns the North-East location of the region (top-right)
func (r Region) Max() Location {
	return r.max
}

// Center returns the mid point location of the region
func (r Region) Center() Location {
	return NewLocation((r.min.lat+r.max.lat)/2, (r.min.lon+r.max.lon)/2)
}

// Encode a latitude/longitude pair into a geohash with the given precision.
func Encode(latitude, longitude float64, precision int) string {

	minLatitude, maxLatitude := -90.0, 90.0
	minLongitude, maxLongitude := -180.0, 180.0
	char, bit := 0, 0
	// Even starts with longitude and toggles with each cycle
	even := true

	var geohash bytes.Buffer
	for geohash.Len() < precision {
		if even { // LONGITUDE
			mid := (minLongitude + maxLongitude) / 2
			if longitude > mid {	// EAST
				char |= bits[bit]
				minLongitude = mid
			} else {				// WEST
				maxLongitude = mid
			}
		} else { // LATITUDE
			mid := (minLatitude + maxLatitude) / 2
			if latitude > mid {		// NORTH
				char |= bits[bit]
				minLatitude = mid
			} else {				//SOUTH
				maxLatitude = mid
			}
		}
		even = !even

		if bit < 4 {
			bit++
		} else {
			geohash.WriteByte(base32[char])
			char, bit = 0, 0
		}
	}
	return geohash.String()
}

// Decode a geohash into a region
func Decode(geohash string) Region {

	minLatitude, maxLatitude := -90.0, 90.0
	minLongitude, maxLongitude := -180.0, 180.0
	// Even starts with longitude and toggles with each cycle
	even := true
	// Iterate over the geohash in byte form, c is each char/byte
	for _, char := range []byte(geohash) {
		// decimal will be the base32-unencoded integer value of char [0-31]
		decimal := bytes.IndexByte(base32, char)
		for i := 0; i < 5; i++ {
			mask := bits[i]
			if even { // longitude
				if decimal&mask != 0 {
					minLongitude = (minLongitude + maxLongitude) / 2	// EAST
				} else {
					maxLongitude = (minLongitude + maxLongitude) / 2	// WEST
				}
			} else { // latitude
				if decimal&mask != 0 {
					minLatitude = (minLatitude + maxLatitude) / 2		// NORTH
				} else {
					maxLatitude = (minLatitude + maxLatitude) / 2		// SOUTH
				}
			}
			even = !even
		}
	}
	return NewRegion(NewLocation(minLatitude, minLongitude), NewLocation(maxLatitude, maxLongitude))
}

// GetNeighbors returns a map of the 8 adjacent neighbouring geohashes of the given geohash within the GeoHasher region.
func GetNeighbors(geohash string) map[string]string {
	region := Decode(geohash)
	/// width and height are deltas for calculating the neighbours
	width := math.Abs(region.Max().Longitude() - region.Min().Longitude())
	height := math.Abs(region.Max().Latitude() - region.Min().Latitude())
	latitude := region.Center().Latitude()
	longitude := region.Center().Longitude()
	precision := len(geohash)
	
	return map[string]string{
		"n" : Encode(latitude+height, longitude, precision),
		"s" : Encode(latitude-height, longitude, precision),
		"e" : Encode(latitude, longitude+width, precision),
		"w" : Encode(latitude, longitude-width, precision),
		"ne": Encode(latitude+height, longitude+width, precision),
		"se": Encode(latitude-height, longitude+width, precision),
		"sw": Encode(latitude-height, longitude-width, precision),
		"nw": Encode(latitude+height, longitude-width, precision),
	}
}

// Valid checks if all the characters in a geohash are valid base32/geohash characters
func Valid(geohash string) bool {
	for _, c := range []byte(geohash) {
		if i := bytes.IndexByte(base32, c); i == -1 {
			return false
		}
	}
	return true
}
