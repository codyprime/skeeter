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
package dummy

import (
	"fmt"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Module struct {
	value  int
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

func (m *Module) MQTTHandler(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("Dummy-MQTT rx: %s: %s\n", msg.Topic(), msg.Payload)
}
