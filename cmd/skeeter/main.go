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
package main

import (
	"flag"
	_ "github.com/codyprime/skeeter/internal/pkg/interfaces"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

type Module struct {
	Class   string           `yaml:"class"`
	Devices []skeeter.Device `yaml:"device"`
}

type Config struct {
	Modules []Module `yaml:"modules"`
}

var conf Config

func main() {
	ex := make(chan os.Signal, 1)
	signal.Notify(ex, os.Interrupt, syscall.SIGTERM)

	username, err := user.Current()
	if err != nil {
		panic(err)
	}

	homedir := username.HomeDir

	defaultPrefsFile := filepath.Join(homedir, ".skeeter.yaml")
	config := flag.String("config", defaultPrefsFile, "configuration file")

	// The rest of our options concern the MQTT broker, and are not contained in
	// any config file.
	// TODO: Use an option config file for MQTT broker
	mqttQos := flag.Int("qos", 1, "MQTT QoS value")
	mqttServer := flag.String("broker", "tcp://192.168.15.2:1883", "MQTT broker")
	mqttRetained := flag.Bool("retained", false, "MQTT broker retains last message")
	mqttUsername := flag.String("username", "", "MQTT broker username")
	mqttPassword := flag.String("password", "", "MQTT broker password")
	mqttLogLevel := flag.String("verbosity", "errors", "Verbosity level: debug, info, errors")

	flag.Parse()

	switch *mqttLogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "errors":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}

	configFile, err := os.Open(*config)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	} else {
		defer configFile.Close()
	}

	prefs, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	err = yaml.Unmarshal([]byte(prefs), &conf)
	if err != nil {
		log.Error(err)
		os.Exit(2)
	}

	mqttOpts := skeeter.MQTTOpts{
		Server:   *mqttServer,
		Clientid: "skeeter-" + strconv.FormatInt(time.Now().Unix(), 16),
		Qos:      byte(*mqttQos),
		Retained: *mqttRetained,
		Username: *mqttUsername,
		Password: *mqttPassword,
	}

	// Go through the device config file, and add each device
	// to the module package.  Every module in conf.Modules needs
	// to have a corresponding implementation under internal/pkg/interfaces
	skeet := skeeter.Skeeter{MQTT: &mqttOpts}
	for _, mod := range conf.Modules {
		modName := mod.Class
		for _, dev := range mod.Devices {
			err := skeet.ModuleAddDevice(modName, &dev)
			if err != nil {
				log.Error(err)
				log.Errorf("Check '%s' for errors\n", *config)
				os.Exit(3)
			}
		}
		//skeeter.ModTest(modName)
	}

	// This also connects to the MQTT broker
	mqttOpts.Init()

	<-ex
}
