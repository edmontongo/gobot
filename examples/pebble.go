package main

import (
	"fmt"

	"github.com/edmontongo/gobot"
	"github.com/edmontongo/gobot/api"
	"github.com/edmontongo/gobot/platforms/pebble"
)

func main() {
	gbot := gobot.NewGobot()
	api := api.NewAPI(gbot)
	api.Port = "8080"
	api.Start()

	pebbleAdaptor := pebble.NewPebbleAdaptor("pebble")
	pebbleDriver := pebble.NewPebbleDriver(pebbleAdaptor, "pebble")

	work := func() {
		pebbleDriver.SendNotification("Hello Pebble!")
		gobot.On(pebbleDriver.Event("button"), func(data interface{}) {
			fmt.Println("Button pushed: " + data.(string))
		})

		gobot.On(pebbleDriver.Event("tap"), func(data interface{}) {
			fmt.Println("Tap event detected")
		})
	}

	robot := gobot.NewRobot("pebble",
		[]gobot.Connection{pebbleAdaptor},
		[]gobot.Device{pebbleDriver},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
