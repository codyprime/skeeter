package main

import (
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/codyprime/skeeter/internal/pkg/interfaces"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
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

	defaultPrefsFile := filepath.Join(homedir, ".skeeter.json")

	config := flag.String("config", defaultPrefsFile, "configuration file")
	mqttQos := flag.Int("qos", 2, "MQTT QoS value")
	mqttServer := flag.String("server", "tcp://192.168.15.2:1883", "MQTT broker")
	mqttRetained := flag.Bool("retained", true, "MQTT broker retains last message")
	mqttUsername := flag.String("username", "", "MQTT broker username")
	mqttPassword := flag.String("password", "", "MQTT broker password")

	flag.Parse()

	configFile, err := os.Open(*config)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		defer configFile.Close()
	}

	prefs, err := ioutil.ReadAll(configFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(prefs))
	err = json.Unmarshal([]byte(prefs), &conf)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(conf)

	mqttOpts := skeeter.MQTTOpts{
		Server:   *mqttServer,
		Clientid: "skeeter-"+strconv.FormatInt(time.Now().Unix(),16),
		Qos:      byte(*mqttQos),
		Retained: *mqttRetained,
		Username: *mqttUsername,
		Password: *mqttPassword,
	}

	mqttOpts.Init()

	skeet := skeeter.Skeeter{MQTT: &mqttOpts}
	for _, mod := range conf.Modules {
		modName := mod.Class
		fmt.Printf("Adding device to '%s'\n", modName)
		for _, dev := range mod.Devices {
			skeet.ModuleAddDevice(modName, &dev)
		}
		skeeter.ModTest(modName)
	}

	<-ex
}
