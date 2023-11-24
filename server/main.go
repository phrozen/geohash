package main

import (
	"os"
	"os/signal"
)

func main() {
	app := NewApp(NewBoltDB("geohash"))
	app.Configure()
	defer app.Shutdown()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	app.Start(stop)
}
