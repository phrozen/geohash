package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/phrozen/geohash"
	bolt "go.etcd.io/bbolt"
)

type server struct {
	db *bolt.DB
}

func (app *server) set(bucket, geohash, data string) error {
	return app.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		err = b.Put([]byte(geohash), []byte(data))
		if err != nil {
			return err
		}
		return nil
	})
}

func (app *server) get(bucket, geohash string) string {
	var data []byte
	app.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		data = b.Get([]byte(geohash))
		return nil
	})
	if data == nil {
		return ""
	}
	return string(data)
}

func (app *server) getPrefix(bucket, geohash string) map[string]string {
	region := make(map[string]string)
	app.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		prefix := []byte(geohash)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			region[string(k)] = string(v)
		}
		return nil
	})
	return region
}

func (app *server) getData(c echo.Context) error {
	data := app.get(c.Param("bucket"), c.Param("geohash"))
	if data == "" {
		return echo.NewHTTPError(http.StatusNotFound, "Geohash not found in bucket")
	}
	return c.String(http.StatusOK, string(data))
}

func (app *server) getAllData(c echo.Context) error {
	data := app.getPrefix(c.Param("bucket"), c.Param("geohash"))
	return c.JSON(http.StatusOK, data)
}

func (app *server) postData(c echo.Context) error {
	data, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(data) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Body must have non-zero length")
	}
	app.set(c.Param("bucket"), c.Param("geohash"), string(data))
	return c.String(http.StatusCreated, fmt.Sprintf("%s/%s", c.Param("bucket"), c.Param("geohash")))
}

// ValidateGeohash is a MiddlewareFunc that checks that the given geohash URL parameter is valid
func ValidateGeohash(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !geohash.Valid(c.Param(("geohash"))) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid character in geohash (base32)")
		}
		return next(c)
	}
}

func main() {
	app := new(server)
	db, err := bolt.Open("geohash.db", 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}
	app.db = db
	defer app.db.Close()

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost},
	}))

	e.Use(ValidateGeohash)

	e.GET("/:bucket/:geohash", app.getData)
	e.POST("/:bucket/:geohash", app.postData)
	e.GET("/:bucket/:geohash/all", app.getAllData)
	e.Logger.Fatal(e.Start(":3000"))

}
