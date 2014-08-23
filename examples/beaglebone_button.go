package main

import (
	"fmt"

	"github.com/edmontongo/gobot"
	"github.com/edmontongo/gobot/platforms/beaglebone"
	"github.com/edmontongo/gobot/platforms/gpio"
)

func main() {
	gbot := gobot.NewGobot()

	beagleboneAdaptor := beaglebone.NewBeagleboneAdaptor("beaglebone")
	button := gpio.NewButtonDriver(beagleboneAdaptor, "button", "P8_9")

	work := func() {
		gobot.On(button.Event("push"), func(data interface{}) {
			fmt.Println("button pressed")
		})

		gobot.On(button.Event("release"), func(data interface{}) {
			fmt.Println("button released")
		})
	}

	robot := gobot.NewRobot("buttonBot",
		[]gobot.Connection{beagleboneAdaptor},
		[]gobot.Device{button},
		work,
	)
	gbot.AddRobot(robot)
	gbot.Start()
}
