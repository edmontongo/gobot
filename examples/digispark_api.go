package main

import (
	"github.com/edmontongo/gobot"
	"github.com/edmontongo/gobot/api"
	"github.com/edmontongo/gobot/platforms/digispark"
	"github.com/edmontongo/gobot/platforms/gpio"
)

func main() {
	gbot := gobot.NewGobot()

	api.NewAPI(gbot).Start()

	digisparkAdaptor := digispark.NewDigisparkAdaptor("Digispark")
	led := gpio.NewLedDriver(digisparkAdaptor, "led", "0")

	robot := gobot.NewRobot("digispark",
		[]gobot.Connection{digisparkAdaptor},
		[]gobot.Device{led},
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
