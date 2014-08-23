package main

import (
	"time"

	"github.com/edmontongo/gobot"
	"github.com/edmontongo/gobot/platforms/firmata"
	"github.com/edmontongo/gobot/platforms/gpio"
)

func main() {
	gbot := gobot.NewGobot()

	firmataAdaptor := firmata.NewFirmataAdaptor("firmata", "/dev/ttyACM0")
	motor := gpio.NewMotorDriver(firmataAdaptor, "motor", "3")

	work := func() {
		speed := byte(0)
		fadeAmount := byte(15)

		gobot.Every(100*time.Millisecond, func() {
			motor.Speed(speed)
			speed = speed + fadeAmount
			if speed == 0 || speed == 255 {
				fadeAmount = -fadeAmount
			}
		})
	}

	robot := gobot.NewRobot("motorBot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{motor},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
