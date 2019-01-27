package main

import (
	_ "github.com/codyprime/skeeter/internal/pkg/interfaces"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
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
	skeeter.ModTest(*device)
}
