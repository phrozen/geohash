package geohash

import (
	"bytes"
	"errors"
)

var (
	//Base32 is the dictionary of characters for generating hashes
	base32 = []byte("0123456789bcdefghjkmnpqrstuvwxyz")
	// Bit positions for 5 bit base32 encoding
	bits = []int{16, 8, 4, 2, 1}
)

// Location is a coordinate pair of latitude, longitude (x, y)
type Location struct {
	lat, lon float64
}

// Latitude returns the latitude of the given Location
func (loc Location) Latitude() float64 {
	return loc.lat
}

// Longitude returns the longitude of the given Location
func (loc Location) Longitude() float64 {
	return loc.lon
}

// NewLocation creates a new location (point) with the given coordinates
func NewLocation(latitude, longitude float64) Location {
	return Location{lat: latitude, lon: longitude}
}

// Region is a bounding box representation of a given area
type Region struct {
	min, max Location
}

// NewRegion ...
func NewRegion(min, max Location) Region {
	return Region{min: min, max: max}
}

// Min ...
func (r Region) Min() Location {
	return r.min
}

// Max ...
func (r Region) Max() Location {
	return r.max
}

// Center returns the mid point of the region
func (r Region) Center() Location {
	return NewLocation((r.min.lat+r.max.lat)/2, (r.min.lon+r.max.lon)/2)
}

// GeoHash ...
type GeoHash struct {
	region    Region
	precision int
}

// NewGeoHash ...
func NewGeoHash(region Region, precision int) *GeoHash {
	return &GeoHash{
		region:    region,
		precision: precision,
	}
}

// Encode ...
func (gh *GeoHash) Encode(latitude, longitude float64) string {

	minLatitude, maxLatitude := gh.region.min.lat, gh.region.max.lat
	minLongitude, maxLongitude := gh.region.min.lon, gh.region.max.lon

	char, bit := 0, 0
	even := true

	var geohash bytes.Buffer
	for geohash.Len() < gh.precision {
		if even {
			mid := (minLongitude + maxLongitude) / 2
			if longitude > mid {
				// EAST
				char |= bits[bit]
				minLongitude = mid
			} else {
				// WEST
				maxLongitude = mid
			}
		} else {
			mid := (minLatitude + maxLatitude) / 2
			if latitude > mid {
				// NORTH
				char |= bits[bit]
				minLatitude = mid
			} else {
				//SOUTH
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

// Decode ...
func (gh *GeoHash) Decode(geohash string) Region {
	minLatitude, maxLatitude := gh.region.min.lat, gh.region.max.lat
	minLongitude, maxLongitude := gh.region.min.lon, gh.region.max.lon
	even := true
	// Iterate over the geohash in byte form, c is each char/byte
	for _, c := range []byte(geohash) {
		d := bytes.IndexByte(base32, c)
		for i := 0; i < 5; i++ {
			mask := bits[i]
			if even {
				if d&mask != 0 {
					// EAST
					minLongitude = (minLongitude + maxLongitude) / 2
				} else {
					// WEST
					maxLongitude = (minLongitude + maxLongitude) / 2
				}
			} else {
				if d&mask != 0 {
					// NORTH
					minLatitude = (minLatitude + maxLatitude) / 2
				} else {
					// SOUTH
					maxLatitude = (minLatitude + maxLatitude) / 2
				}
			}
			even = !even
		}

	}
	return NewRegion(NewLocation(minLatitude, minLongitude), NewLocation(maxLatitude, maxLongitude))
}

// GetNeighbors ...
func (gh *GeoHash) GetNeighbors(geohash string) map[string]string {
	/*
		func GetNeighbors(latitude, longitude float64, precision int) []string {
			geohashs := make([]string, 9)

			// 本身
			geohash, b := Encode(latitude, longitude, precision)
			geohashs[0] = geohash

			// 上下左右
			geohashUp, _ := Encode((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2, precision)
			geohashDown, _ := Encode((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2, precision)
			geohashLeft, _ := Encode((b.MinLat+b.MaxLat)/2, (b.MinLng+b.MaxLng)/2-b.Width(), precision)
			geohashRight, _ := Encode((b.MinLat+b.MaxLat)/2, (b.MinLng+b.MaxLng)/2+b.Width(), precision)

			// 四个角
			geohashLeftUp, _ := Encode((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2-b.Width(), precision)
			geohashLeftDown, _ := Encode((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2-b.Width(), precision)
			geohashRightUp, _ := Encode((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2+b.Width(), precision)
			geohashRightDown, _ := Encode((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2+b.Width(), precision)

			geohashs[1], geohashs[2], geohashs[3], geohashs[4] = geohashUp, geohashDown, geohashLeft, geohashRight
			geohashs[5], geohashs[6], geohashs[7], geohashs[8] = geohashLeftUp, geohashLeftDown, geohashRightUp, geohashRightDown

			return geohashs
		}
	*/
	return make(map[string]string)
}

// The default geohash covers the entire globe with a default precision of 12
// The coordinate region goes from [-90, -180] up to [90, 180]
var globe = NewGeoHash(NewRegion(NewLocation(-90.0, -180.0), NewLocation(90.0, 180.0)), 12)

// From: https://en.wikipedia.org/wiki/Geohash
// The globe covers this errors with the given hash length (precision)
// geohash length	lat bits	lng bits	lat error	lng error	km error
// 			1			2			3		 ±23		  ±23	 	  ±2500
// 			2			5			5		 ±2.8	 	  ±5.6	 	  ±630
// 			3			7			8		 ±0.70	 	  ±0.70	  	  ±78
// 			4			10			10		 ±0.087	 	  ±0.18	  	  ±20
// 			5			12			13		 ±0.022	 	  ±0.022	  ±2.4
// 			6			15			15		 ±0.0027	  ±0.0055	  ±0.61
// 			7			17			18		 ±0.00068	  ±0.00068	  ±0.076
// 			8			20			20		 ±0.000085	  ±0.00017	  ±0.019

// Encode uses the default GeoHash (globe) with a precision of 12
func Encode(lat, lon float64) string {
	return globe.Encode(lat, lon)
}

// Decode uses the default GeoHash (globe) to return a region
func Decode(geohash string) Region {
	return globe.Decode(geohash)
}

// Validate checks if all the characters in a geohash are valid base32 characters
func Validate(geohash string) error {
	for _, c := range []byte(geohash) {
		if i := bytes.IndexByte(base32, c); i == -1 {
			return errors.New("Invalid character in geohash (base32)")
		}
	}
	return nil
}
