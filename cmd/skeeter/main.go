package main

import (
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/codyprime/skeeter/internal/pkg/interfaces"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
	"io/ioutil"
	//"net"
	"os"
	"os/user"
	"path/filepath"
	//"bytes"
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

	username, err := user.Current()
	if err != nil {
		panic(err)
	}

	homedir := username.HomeDir

	defaultPrefsFile := filepath.Join(homedir, ".skeeter.json")

	config := flag.String("config", defaultPrefsFile, "configuration file")

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

	err = json.Unmarshal([]byte(prefs), &conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, mod := range conf.Modules {
		modName := mod.Class
		fmt.Printf("Adding device to '%s'\n", modName)
		for _, dev := range mod.Devices {
			skeeter.ModuleAddDevice(modName, &dev)
		}
		skeeter.ModTest(modName)
	}
}
