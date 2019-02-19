package tplink

import (
	"fmt"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
	"net"
)

func Connect(ip string, port string) (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", ip + ":" + port)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

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

func (m *Module) AddDevice(ip string, port string, id string, devType string) {
	fmt.Printf("Will monitor %s:%s, id %s, %s\n", ip, port, id, devType)

	// TODO: add each device to a map, and in a separate goroutine connect to devices
	conn, err := Connect(ip, port)
	if err != nil {
		fmt.Println(err)
	}

	device := KasaDevice{ Conn: conn }
	go device.KasaComm()
}

func (m *Module) MessageRx(topic string, payload string) {
}
