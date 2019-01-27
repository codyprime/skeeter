package dummy

import (
	"fmt"
	"github.com/codyprime/mqtt-att/mqttatt"
)

type Device struct {
	value int
}

var myDev *Device = &Device{}

func init() {
	mqttatt.Register("dummy", myDev)
}

func (d *Device) ModuleTest() {
	fmt.Println("dummy ModuleTest")
}
