package main

import (
	"encoding/json"
	"flag"
	_ "github.com/codyprime/skeeter/internal/pkg/interfaces"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
	log "github.com/sirupsen/logrus"
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
	Class   string           `json:"class"`
	Devices []skeeter.Device `json:"device"`
}

type Config struct {
	Modules []Module `json:"modules"`
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

	// Right now, the so-called "config file" is just a json file describing
	// each device. See "example-config.json".
	// TODO: Use a different format for configuring devices
	defaultPrefsFile := filepath.Join(homedir, ".skeeter.json")
	config := flag.String("config", defaultPrefsFile, "configuration file")

	// The rest of our options concern the MQTT broker, and are not contained in
	// any config file.
	// TODO: Use an option config file for MQTT broker
	mqttQos := flag.Int("qos", 2, "MQTT QoS value")
	mqttServer := flag.String("broker", "tcp://192.168.15.2:1883", "MQTT broker")
	mqttRetained := flag.Bool("retained", true, "MQTT broker retains last message")
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

	err = json.Unmarshal([]byte(prefs), &conf)
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

	// This also connects to the MQTT broker
	mqttOpts.Init()

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

	<-ex
}
