package dummy

import (
	"fmt"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Module struct {
	value int
	Device skeeter.Device
}

var myDev *Module = &Module{}

func init() {
	skeeter.RegisterModule("dummy", myDev)
}

func (m *Module) ModuleTest() {
	fmt.Println("dummy ModuleTest")
}

func (m *Module) AddDevice(device skeeter.Device, mqtt *skeeter.MQTTOpts, topics []string) {
	m.Device = device
	fmt.Printf("Will monitor %s, id %d, %s\n", device.IP, device.ID, device.Type)
}

func (m *Module) MessageRx(topic string, payload string) {
}

func (m *Module) MQTTHandler(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("Dummy-MQTT rx: %s: %s\n", msg.Topic(), msg.Payload)
}
