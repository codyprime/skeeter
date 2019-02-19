package tplink

import (
	"encoding/binary"
	"encoding/hex"
	//"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

// Every device has its own goroutine for communication

// Comm elements for each device
type KasaDevice struct {
	Conn     net.Conn
	c        chan Cmds
	r        chan MsgResp
}

//========================================================================
// Send and receive responses for any periodic messages
func (k *KasaDevice) kasaPoll() {
	var giResp *SystemResponse
	var ok bool
	for {
		k.c <- CMD_INFO
		mResp := <-k.r

		resp, _ := mResp.Cmd.Unmarshal(mResp.Data[:])
		giResp, ok = resp.(*SystemResponse)
		if ok {
			fmt.Printf("relay state: %d, brightness: %d\n",
				giResp.System.GetSysInfo.RelayState, giResp.System.GetSysInfo.Brightness)
		} else {
			fmt.Println("Not OK :(")
		}
		// TODO: Make this configurable per device
		time.Sleep(500 * time.Millisecond)
	}
}

//========================================================================
// Transmit and Receive messages from the TP-Link device
func (k *KasaDevice) KasaTxRx(txData []byte) (rxLen uint32, rxData []byte, err error) {
	n, err := k.Conn.Write(txData)
	fmt.Printf("Sent %d bytes\n", n)
	if err != nil {
		fmt.Println(err)
		return 0, nil, err
	}
	err = binary.Read(k.Conn, binary.BigEndian, &rxLen)
	if err != nil {
		fmt.Println(err)
		return 0, nil, err
	}
	if rxLen > RX_MAX_LEN {
		fmt.Printf("Response size %d is too large\n", rxLen)
		return 0, nil, err
	}
	rxData = make([]byte, rxLen)
	n, err = io.ReadFull(k.Conn, rxData[:rxLen])
	if err != nil {
		fmt.Println(err)
		rxData = nil	// hint for the gc
		return 0, nil, err
	}
	return rxLen, rxData, nil
}

//========================================================================
// Simple function intended to be run as a separate Goroutine.  Listens
// for a CMD over a channel 'c', and sends a received response back
// over 'r'.
func (k *KasaDevice) KasaComm() (err error) {
	var encMsg []byte
	var msgResp MsgResp

	k.c = make(chan Cmds)
	k.r = make(chan MsgResp)
	// Launch another goroutine to take care of any state polling we
	// wish to do
	go k.kasaPoll()

	for {
		cmd := <-k.c
		encMsg, err = cmd.Marshal()
		if err != nil {
			msgResp.Cmd = CMD_ERR
			k.r <- msgResp
			continue
		}
		fmt.Printf("Sending encrypted infoMsg:\n%s\n", hex.Dump(encMsg))
		fmt.Println(string(kasaDecode(encMsg[4:])))

		rxLen, encData, err := k.KasaTxRx(encMsg[:])
		if err != nil {
			msgResp.Cmd = CMD_ERR
		} else {
			fmt.Printf("Received %d bytes\n", rxLen)
			msgResp.Cmd = cmd
			msgResp.Len = rxLen
			msgResp.Data = encData[:rxLen]
		}
		k.r <- msgResp
	}

	return nil
}
