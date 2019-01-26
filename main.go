package main

import (
	_ "github.com/codyprime/mqtt-att/interfaces"
	"flag"
	"fmt"
	"net"
)

func main() {

	ipStr := flag.String("ip", "127.0.0.1", "ip address of device")

	flag.Parse()

	ip := net.ParseIP(*ipStr)

	fmt.Println(ip)
}
