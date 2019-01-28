package tplink

import (
	"fmt"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
)

type Module struct {
	value int
}

var myDev *Module = &Module{}

func init() {
	skeeter.RegisterModule("tplink", myDev)
}

func (m *Module) ModuleTest() {
	fmt.Println("tplink ModuleTest")
}

func (m *Module) AddDevice(ip string, id string, devType string) {
	fmt.Printf("Will monitor %s, id %s, %s\n", ip, id, devType)
}

func (m *Module) MessageRx(topic string, payload string) {
}
