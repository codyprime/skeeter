package skeeter

import (
	"fmt"
	"net"
)

type Module interface {
	ModuleTest()
}

type Options struct {
	IP net.IP
	ID string // Will be appended to device name
}

type deviceInfo struct {
	module  Module
	devices map[string]Options // map key is the device ID
}

var devices map[string]deviceInfo

func init() {
	devices = make(map[string]deviceInfo)
}

// other module ideas:
//  -- load avg (computer stats)
//  -- presence (arping)
//  -- bandwidth usage
//  -- market data (https://github.com/timpalpant/go-iex)

// compile topic
// pass in topic message rx handler
// maybe pass all this in a struct
func RegisterModule(name string, module Module) (err error) {
	if module == nil {
		err = fmt.Errorf("Cannot register module, nil value")
		fmt.Println(err)
		return err
	}

	devices[name] = deviceInfo{module: module}
	//devices[name].devices = make(map[string]Options)

	return nil
}

// `name` must be the same as used to register the module
func ModuleAddDevice(name string, ip string, id string) (err error) {
	dev, ok := devices[name]
	if !ok {
		err = fmt.Errorf("Module %s has not been registered", name)
		fmt.Println(err)
		return err
	}
	fmt.Println(dev)
	return nil
}

func ModTest(name string) {
	dev, ok := devices[name]
	if !ok || dev.module == nil {
		return
	}
	dev.module.ModuleTest()
}
