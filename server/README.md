# geohash based key/value server
[![Build Status](https://travis-ci.org/phrozen/geohash.svg?branch=master)](https://travis-ci.org/phrozen/geohash)
[![GoDoc](https://godoc.org/github.com/phtozen/geohash?status.svg)](https://godoc.org/github.com/phrozen/geohash)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/phrozen/geohash)](https://goreportcard.com/report/github.com/phrozen/geohash)
[![codecov](https://codecov.io/gh/phrozen/geohash/branch/master/graph/badge.svg)](https://codecov.io/gh/phrozen/geohash)

[![Maintainability](https://api.codeclimate.com/v1/badges/8e62654db4b7c44b0087/maintainability)](https://codeclimate.com/github/phrozen/geohash/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/8e62654db4b7c44b0087/test_coverage)](https://codeclimate.com/github/phrozen/geohash/test_coverage)


### Endpoints
```
GET        /:geohash
POST       /:geohash
GET        /:geohash/region
GET        /:geohash/neighbours
OPTIONS    /:geohash*
```