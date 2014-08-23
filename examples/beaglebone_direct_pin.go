package main

import (
	"time"

	"github.com/edmontongo/gobot"
	"github.com/edmontongo/gobot/platforms/beaglebone"
	"github.com/edmontongo/gobot/platforms/gpio"
)

func main() {
	gbot := gobot.NewGobot()

	beagleboneAdaptor := beaglebone.NewBeagleboneAdaptor("beaglebone")
	led := gpio.NewDirectPinDriver(beagleboneAdaptor, "led", "P8_10")
	button := gpio.NewDirectPinDriver(beagleboneAdaptor, "button", "P8_9")

	work := func() {
		gobot.Every(500*time.Millisecond, func() {
			if button.DigitalRead() == 1 {
				led.DigitalWrite(1)
			} else {
				led.DigitalWrite(0)
			}
		})
	}

	robot := gobot.NewRobot("pinBot",
		[]gobot.Connection{beagleboneAdaptor},
		[]gobot.Device{led},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
