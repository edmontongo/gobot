package main

import (
	"fmt"
	"time"

	"github.com/edmontongo/gobot"
	"github.com/edmontongo/gobot/platforms/sphero"
)

func main() {
	gbot := gobot.NewGobot()

	adaptor := sphero.NewSpheroAdaptor("Sphero", "/dev/rfcomm0")
	spheroDriver := sphero.NewSpheroDriver(adaptor, "sphero")
	collisions := 0
	work := func() {

		spheroDriver.ConfigureCollisionDetectionRaw(0x10, 0x01, 0x10, 0x01, 200)

		gobot.On(spheroDriver.Event("collision"), func(data interface{}) {
			fmt.Printf("Collision Detected!%+v\n", data)
			collisions = collisions + 1
		})

		gobot.On(spheroDriver.Event("locator"), func(data interface{}) {
			fmt.Printf("Locator Detected!%+v\n", data)
		})


		gobot.Every(time.Second, func() {
			// just hit the sphero around
			//spheroDriver.Roll(uint8(gobot.Rand(1)), uint16(gobot.Rand(360)))
			fmt.Printf("Collisions: %v\n", collisions)

		})
		gobot.Every(time.Second, func() {
			// Lame, keep reinstalling streaming
			// and collisions til we're sure we have a collision!
			if (collisions < 1) {
				fmt.Printf("Trying to enable Collision Detection!\n")
				spheroDriver.ConfigureCollisionDetectionRaw(0x20, 0x20, 0x20, 0x20, 200)
				fmt.Printf("Trying to enable LOcator!\n")

				spheroDriver.ConfigureLocatorStreaming(2)
			}
		})

		gobot.Every(1*time.Second, func() {
			r := uint8(255)
			g := uint8(gobot.Rand(255))
			b := uint8(gobot.Rand(255))
			spheroDriver.SetRGB(r, g, b)
		})

	}

	robot := gobot.NewRobot("sphero",
		[]gobot.Connection{adaptor},
		[]gobot.Device{spheroDriver},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
