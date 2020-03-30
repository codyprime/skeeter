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
package skeeter

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

//---------------------------------------------------------
// MQTT topic suffixes to subscribe to.
// The full topics will be of the form:
//   skeeter/[modulename]/[type]-[id]/[suffix]
//
//
// Example:
//   name/Class: "tplink"
//   Device : {
//      ID:   "b0:be:76:a9:ee:0d"
//      Type: "switch"
//      Subs: [ "suffix": "state", "suffix": "brightness" ]
//
//  These topics will be sent to the module to handle:
//		skeeter/tplink/switch/b0:be:76:a9:ee:0d/state
//		skeeter/tplink/switch/b0:be:76:a9:ee:0d/brightness
//

type Device struct {
	IP     string   `yaml:"ip"`
	Port   string   `yaml:"port"`
	ID     string   `yaml:"id"`
	Type   string   `yaml:"type"`
	Subs   []string `yaml:"sub-suffixes"`
	Pubs   []string `yaml:"pub-suffixes"`
	PollMs int      `yaml:"poll-interval"`
}

type Module interface {
	ModuleTest()
	// Add a device to monitor
	AddDevice(device Device, mqtt *MQTTOpts, topics []string)

	MQTTHandler(MQTT.Client, MQTT.Message)
}

type Skeeter struct {
	MQTT *MQTTOpts
}

type deviceInfo struct {
	module  Module
	devices map[string]*Device // map key is the device ID
}

var devInfo map[string]deviceInfo

func init() {
	devInfo = make(map[string]deviceInfo)
}

//========================================================================
// Register a module.
//
// This is expected to be called in a module's init() function.
//
// This will create a map entry for the module name,, and initialize
// its devices map.  Must be called prior to any callbacks.
func RegisterModule(name string, module Module) (err error) {
	if module == nil {
		err = fmt.Errorf("Cannot register module, nil value")
		return err
	}

	_, ok := devInfo[name]
	if ok {
		log.Errorf("Module %s already registered\n", name)
	} else {
		devInfo[name] = deviceInfo{
			module:  module,
			devices: make(map[string]*Device),
		}
	}
	return nil
}

//========================================================================
// Add a device to a module
//
// The passed name must match a name of an existing module
func (s *Skeeter) ModuleAddDevice(name string, dev *Device) (err error) {
	devlist, ok := devInfo[name]
	if !ok {
		err = fmt.Errorf("Module %s has not been registered, or does not exist", name)
		return err
	}
	if dev == nil {
		err = fmt.Errorf("*Device cannot be nil")
		return err
	}

	subTopics := make([]string, len(dev.Subs))
	for i, topic := range dev.Subs {
		subTopics[i] = name + "/" + dev.Type + "/" + dev.ID + "/" + topic
	}
	devlist.devices[dev.ID] = dev
	devlist.module.AddDevice(*dev, s.MQTT, subTopics)
	// Do AddSubscription after module.AddDevice to avoid race conditions
	// from spawned MQTT goroutine
	for _, topic := range subTopics {
		s.MQTT.AddSubscription(topic, devlist.module.MQTTHandler)
	}
	return nil
}

func ModTest(name string) {
	devlist, ok := devInfo[name]
	if !ok || devlist.module == nil {
		return
	}
	devlist.module.ModuleTest()
}
