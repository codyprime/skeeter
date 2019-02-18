package tplink

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
	"io"
//	"io/ioutil"
	"net"
	"time"
)

func Connect(ip string, port string) (conn net.Conn, err error) {
	conn, err = net.Dial("tcp", ip + ":" + port)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// See: https://github.com/softScheck/tplink-smartplug/blob/master/tplink_smartplug.py
func kasaEncode(msg []byte) (enc []byte) {
	var encBuffer bytes.Buffer
	//enc = make([]byte, len(msg)+4)
	binary.Write(&encBuffer, binary.BigEndian, uint32(len(msg)))

	key := byte(171)
	for i := 0; i < len(msg); i++ {
		e := msg[i] ^ key
		key = e
		_ = encBuffer.WriteByte(e)
	}
	return encBuffer.Bytes()
}

func kasaDecode(enc []byte) (msg []byte) {
	msg = make([]byte, len(enc))

	key := byte(171)
	for i := 0; i < len(enc); i++ {
		msg[i] = enc[i] ^ key
		key = enc[i]
	}
	return msg
}

const RX_MAX_LEN = 4096
var infoMsg = []byte("{\"system\":{\"get_sysinfo\":{}}}")

func KasaComm(conn net.Conn, c chan []byte) (err error) {
	encInfoMsg := kasaEncode(infoMsg)
	var rxLen uint32
	encResp := make([]byte, 4096)
	for {
		fmt.Printf("Sending encrypted infoMsg:\n%s\n", hex.Dump(encInfoMsg))
		fmt.Println(string(kasaDecode(encInfoMsg[4:])))

		n, err := conn.Write(encInfoMsg)
		fmt.Printf("Sent %d bytes\n", n)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Reading encrypted result")
			err = binary.Read(conn, binary.BigEndian, &rxLen)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if rxLen > RX_MAX_LEN {
				fmt.Printf("Response size %d is too large\n", rxLen)
				continue
			}
			n, err = io.ReadFull(conn, encResp[:rxLen])
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("Received %d bytes\n", n)
				if len(encResp) > 4 {
					resp := kasaDecode(encResp[4:rxLen])
					fmt.Printf("%s\n", resp)
				}
			}
		}

		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

type Module struct {
	value int
}

var myDev *Module = &Module{}

func init() {
	skeeter.RegisterModule("tplink", myDev)
}

func (m *Module) ModuleTest() {
	fmt.Println("tplink ModuleTest")
}

func (m *Module) AddDevice(ip string, port string, id string, devType string) {
	fmt.Printf("Will monitor %s:%s, id %s, %s\n", ip, port, id, devType)

	// TODO: add each device to a map, and in a separate goroutine connect to devices
	conn, err := Connect(ip, port)
	if err != nil {
		fmt.Println(err)
	}

	c := make(chan []byte)
	go KasaComm(conn, c)
}

func (m *Module) MessageRx(topic string, payload string) {
}
