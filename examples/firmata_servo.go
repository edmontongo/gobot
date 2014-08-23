package main

import (
	"fmt"
	"time"

	"github.com/edmontongo/gobot"
	"github.com/edmontongo/gobot/platforms/firmata"
	"github.com/edmontongo/gobot/platforms/gpio"
)

func main() {
	gbot := gobot.NewGobot()

	firmataAdaptor := firmata.NewFirmataAdaptor("firmata", "/dev/ttyACM0")
	servo := gpio.NewServoDriver(firmataAdaptor, "servo", "3")

	work := func() {
		gobot.Every(1*time.Second, func() {
			i := uint8(gobot.Rand(180))
			fmt.Println("Turning", i)
			servo.Move(i)
		})
	}

	robot := gobot.NewRobot("servoBot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{servo},
		work,
	)

	gbot.AddRobot(robot)
	gbot.Start()
}
