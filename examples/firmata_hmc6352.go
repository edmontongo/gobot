package main

import (
	"fmt"
	"time"

	"github.com/edmontongo/gobot"
	"github.com/edmontongo/gobot/platforms/firmata"
	"github.com/edmontongo/gobot/platforms/i2c"
)

func main() {
	gbot := gobot.NewGobot()

	firmataAdaptor := firmata.NewFirmataAdaptor("firmata", "/dev/ttyACM0")
	hmc6352 := i2c.NewHMC6352Driver(firmataAdaptor, "hmc6352")

	work := func() {
		gobot.Every(100*time.Millisecond, func() {
			fmt.Println("Heading", hmc6352.Heading)
		})
	}

	robot := gobot.NewRobot("hmc6352Bot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{hmc6352},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
