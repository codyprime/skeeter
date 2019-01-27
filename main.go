package main

import (
	_ "github.com/codyprime/mqtt-att/interfaces"
	"github.com/codyprime/mqtt-att/mqttatt"
	"flag"
	"fmt"
	"net"
)

func main() {

	ipStr := flag.String("ip", "127.0.0.1", "ip address of device")
	device := flag.String("device", "dummy", "device type to wrap")

	flag.Parse()

	ip := net.ParseIP(*ipStr)

	fmt.Println(ip)
	mqttatt.ModTest(*device)
}
