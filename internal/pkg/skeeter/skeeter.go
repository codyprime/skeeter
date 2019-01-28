package skeeter

import (
	"fmt"
	//	"net"
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
//		skeeter/tplink/switch-b0:be:76:a9:ee:0d/state
//		skeeter/tplink/switch-b0:be:76:a9:ee:0d/brightness
//
type SubPub struct {
	Suffix string `json:"suffix"`
}

type Device struct {
	IP   string   `json:"ip"`
	ID   string   `json:"id"`
	Type string   `json:"type"`
	Subs []SubPub `json:"sub_suffixes"`
	Pubs []SubPub `json:"sub_suffixes"`
}

type Module interface {
	ModuleTest()
	// Add a device to monitor
	AddDevice(ip string, id string, devType string)

	MessageRx(topic string, payload string) // handler for subscribed topics
}

type deviceInfo struct {
	module  Module
	devices map[string]*Device // map key is the device ID
}

var devInfo map[string]deviceInfo

func init() {
	devInfo = make(map[string]deviceInfo)
}

// other module ideas:
//  -- load avg (computer stats)
//  -- presence (arping)
//  -- bandwidth usage
//  -- market data (https://github.com/timpalpant/go-iex)

//----------------------------------------------------------------------------
// Register a module.
//
// This is expected to be called in a module's init() function.
//
// This will create a map entry for the module name,, and initialize
// its devices map.  Must be called prior to any callbacks.
func RegisterModule(name string, module Module) (err error) {
	if module == nil {
		err = fmt.Errorf("Cannot register module, nil value")
		fmt.Println(err)
		return err
	}

	_, ok := devInfo[name]
	if ok {
		fmt.Printf("Module %s already registered\n", name)
	} else {
		devInfo[name] = deviceInfo{
			module:  module,
			devices: make(map[string]*Device),
		}
	}
	return nil
}

//----------------------------------------------------------------------------
// Add a device to a module
//
// The passed name must match a name of an existing module
//
func ModuleAddDevice(name string, dev *Device) (err error) {
	devlist, ok := devInfo[name]
	if !ok {
		err = fmt.Errorf("Module %s has not been registered", name)
		fmt.Println(err)
		return err
	}
	if dev == nil {
		err = fmt.Errorf("*Device cannot be nil")
		fmt.Println(err)
		return err
	}

	devlist.devices[name] = dev
	devlist.module.AddDevice(dev.IP, dev.ID, dev.Type)
	fmt.Println(devlist)
	return nil
}

func ModTest(name string) {
	devlist, ok := devInfo[name]
	if !ok || devlist.module == nil {
		return
	}
	devlist.module.ModuleTest()
}
