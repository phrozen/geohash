package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/phrozen/geohash"
)

// App type is to store dependency data
type App struct {
	DB   Database
	echo *echo.Echo
	port string
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
	app.echo = echo.New()
	app.port = "1985"
	if os.Getenv("PORT") != "" {
		app.port = os.Getenv("PORT")
	}
	return app
}

// Shutdown executes cleanup code for the app to gracefully shutdown
func (app *App) Shutdown() {
	err := app.DB.Close()
	if err != nil {
		log.Fatalf("Error closing the database: %v", err)
	}
	log.Println("Database closed")
	log.Println("Server shutdown gracefully")
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

// Configure sets all middleware chains and routes
func (app *App) Configure() {
	// Middleware
	app.echo.Use(middleware.Logger())
	app.echo.Use(middleware.Recover())
	app.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost},
	}))
	app.echo.Use(ValidateGeohash)

	//Routes
	app.echo.GET("/:geohash", app.getDataHandler)
	app.echo.POST("/:geohash", app.postDataHandler)
	app.echo.GET("/:geohash/region", app.getRegionDataHandler)
	app.echo.GET("/:geohash/neighbours", app.getNeighboursDataHandler)
}

// Start the server on a separate goroutine and block until quit signal received
func (app *App) Start(stop chan os.Signal) error {
	go func() {
		if err := app.echo.Start(":" + app.port); err != nil {
			log.Println("Shutting down the server")
		}
	}()

	<-stop // Blocks until signal received in channel
	log.Println("Received OS shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.echo.Shutdown(ctx); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func main() {
	app := NewApp(NewBoltDB("geohash"))
	app.Configure()
	defer app.Shutdown()

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	app.Start(stop)
}
