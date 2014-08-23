package i2c

import (
	"github.com/edmontongo/gobot"
	"testing"
)

func initTestBlinkMDriver() *BlinkMDriver {
	return NewBlinkMDriver(newI2cTestAdaptor("adaptor"), "bot")
}

func TestBlinkMDriverStart(t *testing.T) {
	d := initTestBlinkMDriver()
	gobot.Assert(t, d.Start(), true)
}
