package tplink

import (
	"fmt"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"net"
	"strings"
)

type Module struct {
	value int
	Conn  net.Conn
	kasas map[string]*KasaDevice
}

var myDev *Module = &Module{}

func init() {
	myDev.kasas = make(map[string]*KasaDevice)
	skeeter.RegisterModule("tplink", myDev)
}

func (m *Module) ModuleTest() {
	fmt.Println("tplink ModuleTest")
}

func (m *Module) AddDevice(device skeeter.Device, mqtt *skeeter.MQTTOpts,
	topics []string) {
	fmt.Printf("Will monitor %s:%s, id %s, %s\n", device.IP, device.Port,
		device.ID, device.Type)

	// TODO: add each device to a map, and in a separate goroutine connect to devices

	kasa := KasaDevice{Device: device, MQTT: mqtt}

	for _, topic := range topics {
		m.kasas[topic] = &kasa
	}

	go kasa.KasaComm()
}

func (m *Module) MessageRx(topic string, payload string) {
}

func (m *Module) MQTTHandler(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TPLink-MQTT rx: %s: %s\n", msg.Topic(), msg.Payload())

	topic := msg.Topic()
	kasa := m.kasas[topic]

	idx := strings.LastIndex(topic, "/")
	if idx < 0 {
		return
	}

	suffix := topic[idx+1:]
	fmt.Printf("topic suffix: %s\n", suffix)
	if kasa != nil {
		match := int(-1)
		var cmd Cmds
		for i, sub := range kasa.Device.Subs {
			fmt.Printf("%s : %s : %d\n", sub, suffix, i)
			if sub == string(suffix) {
				match = i
				fmt.Printf("Match at %d for %s, %s\n", match, sub, suffix)
			}
		}
		switch match {
		case 0:
			cmd = CMD_RELAY
		case 1:
			cmd = CMD_BRIGHTNESS
		}
		kasaMsg := MsgSend { Cmd : cmd, Data : msg.Payload() }
		kasa.SendCmd(kasaMsg)
	}
}
