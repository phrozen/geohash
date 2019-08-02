package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/phrozen/geohash"
	bolt "go.etcd.io/bbolt"
)

const (
	bucket = "geohash"
)

type server struct {
	db *bolt.DB
	bucket []byte
}

func (app *server) set(geohash, data []byte) error {
	return app.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(app.bucket)
		return b.Put([]byte(geohash), []byte(data))
	})
}

func (app *server) get(geohash []byte) string {
	var data bytes.Buffer
	app.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(app.bucket)
		data.Write(b.Get(geohash))
		return nil
	})
	return data.String()
}

func (app *server) getPrefix(geohash []byte) map[string]string {
	region := make(map[string]string)
	app.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(app.bucket).Cursor()
		for k, v := c.Seek(geohash); k != nil && bytes.HasPrefix(k, geohash); k, v = c.Next() {
			region[string(k)] = string(v)
		}
		return nil
	})
	return region
}

func (app *server) getData(c echo.Context) error {
	data := app.get([]byte(c.Param("geohash")))
	if len(data) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "Geohash not found")
	}
	return c.String(http.StatusOK, string(data))
}

func (app *server) postData(c echo.Context) error {
	data, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(data) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Body must have non-zero length")
	}
	app.set([]byte(c.Param("geohash")), data)
	return c.String(http.StatusCreated, c.Param("geohash"))
}

func (app *server) getRegionData(c echo.Context) error {
	data := app.getPrefix([]byte(c.Param("geohash")))
	if len(data) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "No geohashes found within region")
	}
	return c.JSON(http.StatusOK, data)
}

func (app *server) getNeighbourData(c echo.Context) error {
	data := make(map[string]map[string]string)
	for k, v := range geohash.Neighbours(c.Param("geohash")) {
		val := app.getPrefix([]byte(v))
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
	e := echo.New()
	app := new(server)
	
	db, err := bolt.Open("geohash.db", 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}
	app.db = db
	defer app.db.Close()

	app.bucket = []byte(bucket)
	err = app.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(app.bucket)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		e.Logger.Fatal(err)
	}
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost},
	}))
	e.Use(ValidateGeohash)
	//Routes
	e.GET("/:geohash", app.getData)
	e.POST("/:geohash", app.postData)
	e.GET("/:geohash/region", app.getRegionData)
	e.GET("/:geohash/neighbours", app.getNeighbourData)
	e.Logger.Fatal(e.Start(":3000"))

}
