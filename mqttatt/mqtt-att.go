package mqttatt

import (
	//"fmt"
)

type Device interface {
	ModuleTest()
}

var devices map[string] Device

func init() {
	devices = make(map[string]Device)
}

func Register(name string, dev Device) {
	devices[name] = dev
}

func ModTest(name string) {
	dev := devices[name]
	if dev == nil {
		return
	}
	dev.ModuleTest()
}
