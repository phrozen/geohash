# geohash
[![Build Status](https://travis-ci.org/phrozen/geohash.svg?branch=master)](https://travis-ci.org/phrozen/geohash)
[![GoDoc](https://godoc.org/github.com/phtozen/geohash?status.svg)](https://godoc.org/github.com/phrozen/geohash)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/phrozen/geohash)](https://goreportcard.com/report/github.com/phrozen/geohash)
[![codecov](https://codecov.io/gh/phrozen/geohash/branch/master/graph/badge.svg)](https://codecov.io/gh/phrozen/geohash)

[![Maintainability](https://api.codeclimate.com/v1/badges/8e62654db4b7c44b0087/maintainability)](https://codeclimate.com/github/phrozen/geohash/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/8e62654db4b7c44b0087/test_coverage)](https://codeclimate.com/github/phrozen/geohash/test_coverage)

Simple implementation of Gustavo Niemeyer's algorithm for tech talk.

> Geohash is a public domain geocode system invented in 2008 by Gustavo Niemeyer[1], which encodes a geographic location into a short string of letters and digits. It is a hierarchical spatial data structure which subdivides space into buckets of grid shape, which is one of the many applications of what is known as a Z-order curve, and generally space-filling curves.
>
> *from [Wikipedia](https://en.wikipedia.org/wiki/Geohash)*