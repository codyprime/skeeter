/*
 * Skeeter - IoT to MQTT Bridge
 *
 * Copyright (C) 2019 Jeff Cody <jeff@codyprime.org>
 *
 * This program is free software; you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation; either version 3 of the License, or (at your option) any later
 * version.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
 * PARTICULAR PURPOSE. See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program; if not, see <https://www.gnu.org/licenses>.
 *
 * Additional permission under GNU GPL version 3 section 7
 *
 * If you modify this Program, or any covered work, by linking or combining it
 * with paho.mqtt.golang (https://github.com/eclipse/paho.mqtt.golang) (or a
 * modified version of that library), containing parts covered by the terms of
 * paho.mqtt.golang, the licensors of this Program grant you additional permission
 * to convey the resulting work.  {Corresponding Source for a non-source form of
 * such a combination shall include the source code for the parts of
 * paho.mqtt.golang  used as well as that of the covered work.}
 */
package tplink

import (
	"fmt"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"math"
	"net"
	"strconv"
	"strings"
	"time"
)

// Helper func for int abs
func abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

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

	log.Infof("Polling tplink device %s every %dms\n", device.ID, device.PollMs)

	kasa := KasaDevice{Device: device, MQTT: mqtt}

	for _, topic := range topics {
		m.kasas[topic] = &kasa
	}

	go kasa.KasaComm()
}

const T_TIME = 6
const MIN_PERIOD = 400

//========================================================================
// dimmerTransition()
//
// Dims the light smoothly over a transition time (T_TIME).
// Will take the light from the current brightness to endBrightness, in
// equal steps over T_TIME seconds.  Minimum sleep period is bounded
// by MIN_PERIOD.
func dimmerTransition(kasa *KasaDevice, endBrightness int) {
	kasa.mux.Lock()
	defer kasa.mux.Unlock()
	msg := MsgSend{Cmd: CMD_BRIGHTNESS}

	if kasa.State == nil {
		log.Errorf("kasa.State == nil\n")
		return
	}

	diff := endBrightness - kasa.State.brightness
	if diff == 0 {
		return
	}

	period := (T_TIME * 1000) / abs(int64(diff))
	b := kasa.State.brightness

	inc := 1
	if diff < 0 {
		inc = -1
	}

	step := 1
	if period < MIN_PERIOD {
		step = int(math.Round(float64(MIN_PERIOD / float64(period))))
		period = MIN_PERIOD
	}

	log.Debugf("dimmerTransition: sleep for %dms, step is %d\n", period, step)
	numSteps := abs(int64(diff / step))
	for i := numSteps; i > 0; i-- {
		b += inc * step
		msg.Data = []byte(fmt.Sprintf("%d", b))
		kasa.QueueCmd(msg)
		time.Sleep(time.Duration(period) * time.Millisecond)
	}
	msg.Data = []byte(fmt.Sprintf("%d", endBrightness))
	kasa.QueueCmd(msg)
	log.Debugf("dimmerTransition: complete\n")
	return
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
		case 2:
			endBrightness, _ := strconv.Atoi(string(msg.Payload()))
			// This will run in the background, but the operation is protected
			// by a mutex, so multiple transition messages will not conflict
			go dimmerTransition(kasa, endBrightness)
		}
		kasaMsg := MsgSend{Cmd: cmd, Data: msg.Payload()}
		kasa.QueueCmd(kasaMsg)
	}
}
