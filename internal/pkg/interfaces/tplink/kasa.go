/*
 * Skeeter - IoT to MQTT Bridge
 *
 * Copyright (C) 2019 Jeff Cody <jeff@codyprime.org>
 *
 * This program is free software; you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation; either version 3 of the License, or (at your option) any later
 * version.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
 * PARTICULAR PURPOSE. See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program; if not, see <https://www.gnu.org/licenses>.
 *
 * Additional permission under GNU GPL version 3 section 7
 *
 * If you modify this Program, or any covered work, by linking or combining it
 * with paho.mqtt.golang (https://github.com/eclipse/paho.mqtt.golang) (or a
 * modified version of that library), containing parts covered by the terms of
 * paho.mqtt.golang, the licensors of this Program grant you additional permission
 * to convey the resulting work.  {Corresponding Source for a non-source form of
 * such a combination shall include the source code for the parts of
 * paho.mqtt.golang  used as well as that of the covered work.}
 */
package tplink

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
	"time"
)

// Every device has its own goroutine for communication

type KasaState struct {
	relay      int
	brightness int
}

// Comm elements for each device
type KasaDevice struct {
	Conn   net.Conn
	Device skeeter.Device
	MQTT   *skeeter.MQTTOpts
	State  *KasaState
	c      chan MsgSend
	r      chan MsgResp
	mux    sync.Mutex
}

func (k *KasaDevice) QueueCmd(cmd MsgSend) (msgResp MsgResp) {
	log.Debugf("QueueCmd: Channel write cmd: %d\n", cmd.Cmd)
	k.c <- cmd
	log.Debugf("QueueCmd: Sent. Awaiting channel rx\n")
	msgResp = <-k.r
	log.Debugf("QueueCmd: Channel resp read\n")
	return msgResp
}

//========================================================================
// Send and receive responses for any periodic messages
func (k *KasaDevice) kasaPoll() {
	var giResp *SystemResponse
	var ok bool
	i := 0
	cmdInfo := MsgSend{Cmd: CMD_INFO}

	if k.State == nil {
		k.State = &KasaState{}
		k.State.relay = -1
		k.State.brightness = -1
	}
	for {
		mResp := k.QueueCmd(cmdInfo)

		i++
		resp, _ := mResp.Cmd.Unmarshal(mResp.Data[:])
		giResp, ok = resp.(*SystemResponse)
		if ok {
			relay := giResp.System.GetSysInfo.RelayState
			brightness := giResp.System.GetSysInfo.Brightness

			log.Infof("relay state: %d, brightness: %d    (msg %d)\n", relay, brightness, i)

			// TODO: Break out topic generation into common function
			//       Will do that once there is a different config file layout
			if relay != k.State.relay {
				k.State.relay = relay
				topic := "tplink/" + k.Device.Type + "/" + k.Device.ID + "/" + k.Device.Pubs[0]
				payload := fmt.Sprintf("%d", relay)
				k.MQTT.Publish(topic, payload)
			}

			if brightness != k.State.brightness {
				k.State.brightness = brightness
				topic := "tplink/" + k.Device.Type + "/" + k.Device.ID + "/" + k.Device.Pubs[1]
				payload := fmt.Sprintf("%d", brightness)
				k.MQTT.Publish(topic, payload)
			}
		} else {
			log.Error("Kasa response does not match expected\n")
		}
		// TODO: Make this configurable per device
		time.Sleep(time.Duration(k.Device.PollMs) * time.Millisecond)
	}
}

//========================================================================
// Connect(timeoutMs)
//
// Does not return until a successful connection has been made.  On error,
// sleep for timeoutMs, and then retry.
func (k *KasaDevice) Connect(timeoutMs int) {
	var err error
	for {
		log.Debugf("Opening connection to %s:%s", k.Device.IP, k.Device.Port)
		k.Conn, err = net.Dial("tcp", k.Device.IP+":"+k.Device.Port)
		if err == nil {
			break
		}
		log.Errorf("Unable to open tcp connection to %s, retrying in %dms\n",
			k.Device.IP, timeoutMs)
		time.Sleep(time.Duration(timeoutMs) * time.Millisecond)
	}
}

//========================================================================
// Transmit and Receive messages from the TP-Link device
func (k *KasaDevice) KasaTxRx(txData []byte) (rxLen uint32, rxData []byte, err error) {
	_, err = k.Conn.Write(txData)
	if err != nil {
		log.Errorf("KasaTxRx Write error: '%s'\n", err)
		return 0, nil, err
	}
	err = binary.Read(k.Conn, binary.BigEndian, &rxLen)
	if err != nil {
		log.Errorf("KasaTxRx Read rxLen error: '%s'\n", err)
		return 0, nil, err
	}
	if rxLen > RX_MAX_LEN {
		log.Warnf("KasaTxRx: Response size %d is too large\n", rxLen)
		return 0, nil, err
	}
	rxData = make([]byte, rxLen)
	_, err = io.ReadFull(k.Conn, rxData[:rxLen])
	if err != nil {
		log.Errorf("KasaTxRx ReadFull error: '%s'\n", err)
		rxData = nil // hint for the gc
		return 0, nil, err
	}
	return rxLen, rxData, nil
}

//========================================================================
// CmdSend:
func (k *KasaDevice) CmdSend(cmd MsgSend) (resp MsgResp, err error) {
	encMsg, err := cmd.Cmd.Marshal(cmd.Data)
	if err != nil {
		resp.Cmd = CMD_ERR
		return resp, err
	}

	log.Debugf("Sending raw encoded msg:\n%s\n", hex.Dump(encMsg))
	log.Debug(string(kasaDecode(encMsg[4:])))
	rxLen, encData, err := k.KasaTxRx(encMsg[:])
	log.Debugf("Received raw encoded msg:\n%s\n", hex.Dump(encData))
	if err != nil {
		log.Errorf("KasaTxRx error '%s', closing connection\n", err)
		k.Conn.Close()
		time.Sleep(time.Duration(1000) * time.Millisecond)
		// This will not return until a succesful connection has
		// been established
		k.Connect(1000)
		resp.Cmd = CMD_ERR
		return resp, err
	}

	resp.Cmd = cmd.Cmd
	resp.Len = rxLen
	resp.Data = encData[:rxLen]

	return resp, nil
}

//========================================================================
// Simple function intended to be run as a separate Goroutine.  Listens
// for a CMD over a channel 'c', and sends a received response back
// over 'r'.
func (k *KasaDevice) KasaComm() (err error) {
	var msgResp MsgResp

	k.c = make(chan MsgSend)
	k.r = make(chan MsgResp)

	// Does not return until a succesful connection is been established.
	k.Connect(1000)
	// Launch another goroutine to take care of any state polling we
	// wish to do
	go k.kasaPoll()

	for {
		log.Debugf("KasaComm: read channel\n")
		cmd := <-k.c
		msgResp, err = k.CmdSend(cmd)
		if err != nil {
			log.Errorf("k.CmdSend(cmd) error: '%s'\n", err)
		}
		log.Debugf("KasaComm: write resp channel\n")
		k.r <- msgResp
	}

	return nil
}
