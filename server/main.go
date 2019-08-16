package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/phrozen/geohash"
	bolt "go.etcd.io/bbolt"
)

// Database defines a simple Key/Value Store interface
type Database interface {
	Open() error
	Close() error
	Set(key, value string) error
	Get(key string) string
	GetAllByPrefix(prefix string) map[string]string
}

// BoltDB implements Database with a BoltDB backend
type BoltDB struct {
	name string
	bolt *bolt.DB
}

// NewBoltDB creates a new BoltDB handler
func NewBoltDB(name string) *BoltDB {
	return &BoltDB{name: name}
}

// Open creates or opens a new BoltDB database file with the filename: <name>.db
func (db *BoltDB) Open() error {
	// Create/Open database file
	database, err := bolt.Open(db.name+".db", 0600, nil)
	if err != nil {
		return err
	}
	db.bolt = database
	// Create default bucket
	err = db.bolt.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(db.name))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

// Close closes the BoltDB database handler and returns it's error if any
func (db *BoltDB) Close() error {
	return db.bolt.Close()
}

// Set stores data to the given geohash key
func (db *BoltDB) Set(geohash, data string) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.name))
		return b.Put([]byte(geohash), []byte(data))
	})
}

// Get returns the data (if any) stored in the geohash key
func (db *BoltDB) Get(geohash string) string {
	var data bytes.Buffer
	db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.name))
		data.Write(b.Get([]byte(geohash)))
		return nil
	})
	return data.String()
}

// GetAllByPrefix returns all the key/value pairs with the given prefix
func (db *BoltDB) GetAllByPrefix(geohash string) map[string]string {
	region := make(map[string]string)
	db.bolt.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(db.name)).Cursor()
		for k, v := c.Seek([]byte(geohash)); k != nil && bytes.HasPrefix(k, []byte(geohash)); k, v = c.Next() {
			region[string(k)] = string(v)
		}
		return nil
	})
	return region
}

// App type is to store dependency data
type App struct {
	DB Database
}

// NewApp creates a new App and initializes database with default bucket
func NewApp(db Database) *App {
	app := &App{DB: db}
	// Open database
	err := app.DB.Open()
	if err != nil {
		log.Fatalf("Error opening the database: %v", err)
		return nil
	}
	return app
}

// Shutdown executes cleanup code for the app to gracefully shutdown
func (app *App) Shutdown() {
	err := app.DB.Close()
	if err != nil {
		log.Fatalf("Error closing the database: %v", err)
	}
}

// GET /:geohash
func (app *App) getDataHandler(c echo.Context) error {
	data := app.DB.Get(c.Param("geohash"))
	if len(data) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "Geohash not found")
	}
	return c.JSON(http.StatusOK, string(data))
}

// POST /:geohash
func (app *App) postDataHandler(c echo.Context) error {
	data, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(data) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Body must have non-zero length")
	}
	app.DB.Set(c.Param("geohash"), string(data))
	return c.JSON(http.StatusCreated, c.Param("geohash"))
}

// GET /:geohash/region
func (app *App) getRegionDataHandler(c echo.Context) error {
	data := app.DB.GetAllByPrefix(c.Param("geohash"))
	if len(data) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "No geohashes found within region")
	}
	return c.JSON(http.StatusOK, data)
}

// GET /:geohash/neighbours
func (app *App) getNeighboursDataHandler(c echo.Context) error {
	data := make(map[string]map[string]string)
	for k, v := range geohash.Neighbours(c.Param("geohash")) {
		val := app.DB.GetAllByPrefix(v)
		if len(val) > 0 {
			data[k] = val
		}
	}
	if len(data) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "No geohashes found within neighbours")
	}
	return c.JSON(http.StatusOK, data)
}

// ValidateGeohash is a MiddlewareFunc that checks that the given geohash URL parameter is valid
func ValidateGeohash(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !geohash.Valid(c.Param("geohash")) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid character in geohash (base32)")
		}
		return next(c)
	}
}

func main() {
	// New Server (App)
	app := NewApp(NewBoltDB("geohash"))
	defer app.Shutdown()

	e := echo.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost},
	}))
	e.Use(ValidateGeohash)

	//Routes
	e.GET("/:geohash", app.getDataHandler)
	e.POST("/:geohash", app.postDataHandler)
	e.GET("/:geohash/region", app.getRegionDataHandler)
	e.GET("/:geohash/neighbours", app.getNeighboursDataHandler)

	// Set a default port and check env var for override
	port := "1985"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	// Run server
	e.Logger.Fatal(e.Start(":" + port))
}
