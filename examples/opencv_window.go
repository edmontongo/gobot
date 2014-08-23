package main

import (
	"github.com/edmontongo/gobot"
	"github.com/edmontongo/gobot/platforms/opencv"
	cv "github.com/hybridgroup/go-opencv/opencv"
)

func main() {
	gbot := gobot.NewGobot()

	window := opencv.NewWindowDriver("window")
	camera := opencv.NewCameraDriver("camera", 0)

	work := func() {
		gobot.On(camera.Event("frame"), func(data interface{}) {
			window.ShowImage(data.(*cv.IplImage))
		})
	}

	robot := gobot.NewRobot("cameraBot",
		[]gobot.Device{window, camera},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
