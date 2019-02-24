package tplink

import (
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
)

type Module struct {
	Conn  net.Conn
	kasas map[string]*KasaDevice
}

var myDev *Module = &Module{}

func init() {
	myDev.kasas = make(map[string]*KasaDevice)
	skeeter.RegisterModule("tplink", myDev)
}

func (m *Module) ModuleTest() {
	log.Info("tplink ModuleTest")
}

func (m *Module) AddDevice(device skeeter.Device, mqtt *skeeter.MQTTOpts,
	topics []string) {
	log.Infof("Will monitor %s:%s, id %s, %s\n", device.IP, device.Port,
		device.ID, device.Type)

	// TODO: add each device to a map, and in a separate goroutine connect to devices
	if device.PollMs == 0 {
		device.PollMs = 500
	}
	kasa := KasaDevice{Device: device, MQTT: mqtt}

	for _, topic := range topics {
		m.kasas[topic] = &kasa
	}

	go kasa.KasaComm()
}

func (m *Module) MQTTHandler(client MQTT.Client, msg MQTT.Message) {
	log.Debugf("TPLink-MQTT rx: %s: %s\n", msg.Topic(), msg.Payload())

	topic := msg.Topic()
	kasa := m.kasas[topic]

	idx := strings.LastIndex(topic, "/")
	if idx < 0 {
		return
	}

	suffix := topic[idx+1:]
	if kasa != nil {
		match := int(-1)
		var cmd Cmds
		for i, sub := range kasa.Device.Subs {
			if sub == string(suffix) {
				match = i
			}
		}
		switch match {
		case 0:
			cmd = CMD_RELAY
		case 1:
			cmd = CMD_BRIGHTNESS
		}
		kasaMsg := MsgSend{Cmd: cmd, Data: msg.Payload()}
		kasa.SendCmd(kasaMsg)
	}
}
