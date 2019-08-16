package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type MockDB struct {
	db map[string]string
}

func (mock *MockDB) Open() error {
	return nil
}

func (mock *MockDB) Close() error {
	return nil
}

func (mock *MockDB) Set(key, value string) error {
	mock.db[key] = value
	return nil
}

func (mock *MockDB) Get(key string) string {
	return mock.db[key]
}

func (mock *MockDB) GetAllByPrefix(prefix string) map[string]string {
	results := make(map[string]string)
	for k, v := range mock.db {
		if strings.HasPrefix(k, prefix) {
			results[k] = v
		}
	}
	return results
}

var (
	testData = map[string]string{
		"3e4mbr3q2w39": "Chile - Easter Island, Anakena Beach",
		"sr2y7kh9bbfk": "Italy - Vatican, Saint Peter's Basillica",
		"ucfv0j9vp0xz": "Moscow - Red Plaza, Lenin's Monument",
		"r3gx2ux9dg0p": "Sydney - Opera House",
		"9g3w81t7mqpx": "Mexico - CDMX Zócalo",
	}
)

func TestPostData(t *testing.T) {
	app := NewApp(&MockDB{make(map[string]string)})
	defer app.Shutdown()
	e := echo.New()
	for k, v := range testData {
		// Posting Data
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(v))
		ctx := e.NewContext(req, rec)
		ctx.SetPath("/:geohash")
		ctx.SetParamNames("geohash")
		ctx.SetParamValues(k)
		// Assertions
		if assert.NoError(t, app.postDataHandler(ctx), "Should not return an error") {
			assert.Equal(t, http.StatusCreated, rec.Code, "Status should be Created - 201")
			assert.Equal(t, fmt.Sprintf("\"%s\"\n", k), rec.Body.String(), "Return value should match key")
		}
	}
}

func TestPostDataEmpty(t *testing.T) {
	app := NewApp(&MockDB{make(map[string]string)})
	defer app.Shutdown()
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/:geohash")
	ctx.SetParamNames("geohash")
	ctx.SetParamValues("qwerty")
	if err := app.postDataHandler(ctx); assert.Error(t, err) {
		he, ok := err.(*echo.HTTPError)
		if assert.True(t, ok) {
			assert.Equal(t, http.StatusBadRequest, he.Code, "Status should be Bad Request - 400")
			assert.Equal(t, "Body must have non-zero length", he.Message)
		}
	}
}

func TestGetData(t *testing.T) {
	app := NewApp(&MockDB{testData})
	defer app.Shutdown()
	e := echo.New()
	for k, v := range testData {
		// Getting Data
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := e.NewContext(req, rec)
		ctx.SetPath("/:geohash")
		ctx.SetParamNames("geohash")
		ctx.SetParamValues(k)
		// Assertions
		if assert.NoError(t, app.getDataHandler(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code, "Status should be 200")
			assert.Equal(t, fmt.Sprintf("\"%s\"\n", v), rec.Body.String(), "Value should match JSON response")
		}
	}
}

func TestGetDataNotFound(t *testing.T) {
	app := NewApp(&MockDB{testData})
	defer app.Shutdown()
	e := echo.New()
	notFound := []string{
		"12345678",
		"90qwerty",
		"upsdfghj",
		"kzxcvbnm",
	}
	for _, v := range notFound {
		// Getting Data
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(v))
		ctx := e.NewContext(req, rec)
		ctx.SetPath("/:geohash")
		ctx.SetParamNames("geohash")
		ctx.SetParamValues(v)
		// Assertions
		if err := app.getDataHandler(ctx); assert.Error(t, err) {
			he, ok := err.(*echo.HTTPError)
			if assert.True(t, ok) {
				assert.Equal(t, http.StatusNotFound, he.Code, "Status should be Bad Request - 400")
				assert.Equal(t, "Geohash not found", he.Message)
			}
		}
	}
}

func TestInvalidGeohash(t *testing.T) {
	app := NewApp(&MockDB{make(map[string]string)})
	defer app.Shutdown()
	// Setup
	invalid := []string{
		"abcdefgh", // contains 'a'
		"ijk12345", // contains 'i'
		"lmn67890", // contains 'l'
		"opqrstuv", // contains 'o'
		"wxyz?!_#", // contains 'special characters'
		"",         // empty string
	}
	e := echo.New()
	for _, v := range invalid {
		req := httptest.NewRequest(http.MethodGet, "/"+v, nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		ctx.SetPath("/:geohash")
		ctx.SetParamNames("geohash")
		ctx.SetParamValues(v)
		// Assertions
		if err := ValidateGeohash(app.getDataHandler)(ctx); assert.Error(t, err) {
			he, ok := err.(*echo.HTTPError)
			if assert.True(t, ok) {
				assert.Equal(t, http.StatusBadRequest, he.Code, "Status shoul dbe Bad Request - 400")
				assert.Equal(t, "Invalid character in geohash (base32)", he.Message)
			}
		}
	}
}
