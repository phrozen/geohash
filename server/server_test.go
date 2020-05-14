package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

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

// DRYing code, creates a Request and a Response Recorder and sets the geohas to Path context
func CreateContextRecord(method, path, body, geohash string) (*httptest.ResponseRecorder, echo.Context) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	ctx := e.NewContext(req, rec)
	ctx.SetPath(path)
	ctx.SetParamNames("geohash")
	ctx.SetParamValues(geohash)
	return rec, ctx
}

func TestPostData(t *testing.T) {
	test := map[string]string{
		"3e4mbr3q2w39": "Chile - Easter Island, Anakena Beach",
		"sr2y7kh9bbfk": "Italy - Vatican, Saint Peter's Basillica",
		"ucfv0j9vp0xz": "Moscow - Red Plaza, Lenin's Monument",
		"r3gx2ux9dg0p": "Sydney - Opera House",
		"9g3w81t7mqpx": "Mexico - CDMX Zócalo",
	}
	app := NewApp(&MockDB{make(map[string]string)}) // empty database
	defer app.Shutdown()
	for k, v := range test {
		// Posting Data
		rec, ctx := CreateContextRecord(http.MethodPost, "/:geohash", v, k)
		// Assertions
		if assert.NoError(t, app.postDataHandler(ctx), "Should not return an error") {
			assert.Equal(t, http.StatusCreated, rec.Code, "Status should be Created - 201")
			assert.Equal(t, fmt.Sprintf("\"%s\"\n", k), rec.Body.String(), "Return value should match key")
		}
	}
}

func TestPostDataEmpty(t *testing.T) {
	app := NewApp(&MockDB{make(map[string]string)}) // empty database
	defer app.Shutdown()
	_, ctx := CreateContextRecord(http.MethodPost, "/:geohash", "", "qwerty")
	if err := app.postDataHandler(ctx); assert.Error(t, err) {
		he, ok := err.(*echo.HTTPError)
		if assert.True(t, ok) {
			assert.Equal(t, http.StatusBadRequest, he.Code, "Status should be Bad Request - 400")
			assert.Equal(t, "Body must have non-zero length", he.Message)
		}
	}
}

func TestGetDataAndValidateMiddleware(t *testing.T) {
	test := map[string]string{
		"3e4mbr3q2w39": "Chile - Easter Island, Anakena Beach",
		"sr2y7kh9bbfk": "Italy - Vatican, Saint Peter's Basillica",
		"ucfv0j9vp0xz": "Moscow - Red Plaza, Lenin's Monument",
		"r3gx2ux9dg0p": "Sydney - Opera House",
		"9g3w81t7mqpx": "Mexico - CDMX Zócalo",
	}
	app := NewApp(&MockDB{test})
	defer app.Shutdown()
	for k, v := range test {
		// Getting Data
		rec, ctx := CreateContextRecord(http.MethodGet, "/:geohash", "", k)
		// Assertions
		if assert.NoError(t, ValidateGeohash(app.getDataHandler)(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code, "Status should be 200")
			assert.Equal(t, fmt.Sprintf("\"%s\"\n", v), rec.Body.String(), "Value should match JSON response")
		}
	}
}

func TestGetDataNotFound(t *testing.T) {
	app := NewApp(&MockDB{make(map[string]string)}) //Empty database
	defer app.Shutdown()
	notFound := []string{"12345678", "90qwerty", "upsdfghj", "kzxcvbnm"}
	for _, v := range notFound {
		// Getting Data
		_, ctx := CreateContextRecord(http.MethodGet, "/:geohash", "", v)
		// Assertions
		if err := app.getDataHandler(ctx); assert.Error(t, err) {
			he, ok := err.(*echo.HTTPError)
			if assert.True(t, ok) {
				assert.Equal(t, http.StatusNotFound, he.Code, "Status should be Not Found - 404")
				assert.Equal(t, "Geohash not found", he.Message)
			}
		}
	}
}

func TestRegionData(t *testing.T) {
	region := map[string]string{
		"9":     "Precision 1",
		"9e":    "Precision 2a",
		"9ew":   "Precision 3a",
		"9ewm":  "Precision 4a",
		"9ewmq": "Precision 5a",
		"9b":    "Precision 2b",
		"9bn":   "Precision 3b",
		"9bnr":  "Precision 4b",
		"9bnrt": "Precision 5b",
	}
	app := NewApp(&MockDB{region})
	defer app.Shutdown()
	// Getting Valid Region Data
	rec, ctx := CreateContextRecord(http.MethodGet, "/:geohash/region", "", "9")
	// Assertions
	if assert.NoError(t, app.getRegionDataHandler(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code, "Status should be 200")
		data, err := json.Marshal(region)
		assert.Nil(t, err, "Should marshal JSON correctly")
		assert.Equal(t, string(data)+"\n", rec.Body.String(), "Value should match JSON response")
	}
	// Not found in region
	_, ctx = CreateContextRecord(http.MethodGet, "/:geohash/region", "", "9q")
	// Assertions
	if err := app.getRegionDataHandler(ctx); assert.Error(t, err) {
		he, ok := err.(*echo.HTTPError)
		if assert.True(t, ok) {
			assert.Equal(t, http.StatusNotFound, he.Code, "Status should be Not Found - 404")
			assert.Equal(t, "No geohashes found within region", he.Message)
		}
	}
}

func TestNeighboursData(t *testing.T) {
	neighbours := map[string]string{
		"b": "North West",
		"c": "North",
		"f": "North East",
		"8": "West",
		"9": "Center",
		"d": "East",
		"2": "South West",
		"3": "South",
		"6": "South East",
	}
	app := NewApp(&MockDB{neighbours})
	defer app.Shutdown()
	for k := range neighbours {
		rec, ctx := CreateContextRecord(http.MethodGet, "/:geohash/neighbours", "", k)
		// Assertions
		if assert.NoError(t, app.getNeighboursDataHandler(ctx)) {
			assert.Equal(t, http.StatusOK, rec.Code, "Status should be 200")
			//assert.Equal(t, fmt.Sprintf("\"%s\"\n", v), rec.Body.String(), "Value should match JSON response")
		}
	}
	// No neighbours found
	_, ctx := CreateContextRecord(http.MethodGet, "/:geohash/neighbours", "", "00")
	// Assertions
	if err := app.getNeighboursDataHandler(ctx); assert.Error(t, err) {
		he, ok := err.(*echo.HTTPError)
		if assert.True(t, ok) {
			assert.Equal(t, http.StatusNotFound, he.Code, "Status should be Not Found - 404")
			assert.Equal(t, "No geohashes found within neighbours", he.Message)
		}
	}
}

func TestInvalidGeohash(t *testing.T) {
	app := NewApp(&MockDB{make(map[string]string)}) // empty database
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
	for _, v := range invalid {
		_, ctx := CreateContextRecord(http.MethodGet, "/:geohash", "", v)
		// Assertions
		if err := ValidateGeohash(app.getDataHandler)(ctx); assert.Error(t, err) {
			he, ok := err.(*echo.HTTPError)
			if assert.True(t, ok) {
				assert.Equal(t, http.StatusBadRequest, he.Code, "Status should be Bad Request - 400")
				assert.Equal(t, "Invalid character in geohash (base32)", he.Message)
			}
		}
	}
}

func TestAppConfiguration(t *testing.T) {
	app := NewApp(&MockDB{make(map[string]string)})
	// Test default port
	assert.Equal(t, "3000", app.port)
	// Setting new port
	os.Setenv("PORT", "3001")
	app = NewApp(&MockDB{make(map[string]string)})
	assert.Equal(t, "3001", app.port)
	app.Configure()
	// 4 routes defined
	//assert.GreaterOrEqual(t, 4, len(app.echo.Routes()))
}

func TestAppStartAndGracefulShutdown(t *testing.T) {
	app := NewApp(&MockDB{make(map[string]string)})
	stop := make(chan os.Signal)
	go func() {
		log.Println("Sleeping for 4 seconds to test server shutdown")
		time.Sleep(2 * time.Second)
		stop <- os.Interrupt
		time.Sleep(2 * time.Second)
		close(stop)
	}()
	assert.NoError(t, app.Start(stop))
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	code := m.Run()
	os.Exit(code)
}
