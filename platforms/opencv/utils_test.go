package opencv

import (
	"path"
	"runtime"
	"testing"

	"github.com/edmontongo/gobot"
	cv "github.com/hybridgroup/go-opencv/opencv"
)

func TestUtils(t *testing.T) {
	_, currentfile, _, _ := runtime.Caller(0)
	image := cv.LoadImage(path.Join(path.Dir(currentfile), "lena-256x256.jpg"))
	rect := DetectFaces("haarcascade_frontalface_alt.xml", image)
	gobot.Refute(t, len(rect), 0)
	gobot.Refute(t, DrawRectangles(image, rect, 0, 0, 0, 0), nil)
}
