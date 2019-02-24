package tplink

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/codyprime/skeeter/internal/pkg/skeeter"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
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
	p      chan byte
}

func (k *KasaDevice) SendCmd(cmd MsgSend) (msgResp MsgResp) {
	k.c <- cmd
	msgResp = <-k.r
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
		mResp := k.SendCmd(cmdInfo)

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
			log.Error("Kasa Comms Not OK :(")
			break
		}
		// TODO: Make this configurable per device
		time.Sleep(time.Duration(k.Device.PollMs) * time.Millisecond)
	}

	k.p <- 0xff
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
// Simple function intended to be run as a separate Goroutine.  Listens
// for a CMD over a channel 'c', and sends a received response back
// over 'r'.
func (k *KasaDevice) KasaComm() (err error) {
	var encMsg []byte
	var msgResp MsgResp

	k.c = make(chan MsgSend)
	k.r = make(chan MsgResp)
	k.p = make(chan byte)

	for {
		k.Conn, err = net.Dial("tcp", k.Device.IP+":"+k.Device.Port)
		if err != nil {
			log.Errorf("KasaTxRx: Unable to connect, Dial failed: '%s'\n", err)
			return err
		}

		// Launch another goroutine to take care of any state polling we
		// wish to do
		go k.kasaPoll()

		for {
			cmd := <-k.c
			encMsg, err = cmd.Cmd.Marshal(cmd.Data)
			if err != nil {
				msgResp.Cmd = CMD_ERR
				k.r <- msgResp
				continue
			}
			log.Debugf("Sending encrypted infoMsg:\n%s\n", hex.Dump(encMsg))
			log.Debug(string(kasaDecode(encMsg[4:])))

			rxLen, encData, err := k.KasaTxRx(encMsg[:])
			if err != nil {
				msgResp.Cmd = CMD_ERR
				break
			} else {
				log.Debugf("Received %d bytes\n", rxLen)
				msgResp.Cmd = cmd.Cmd
				msgResp.Len = rxLen
				msgResp.Data = encData[:rxLen]
			}
			k.r <- msgResp
		}
		pollExit := <-k.p
		k.Conn.Close()

		// Under normal error, retry establishing connection
		if pollExit != 0xff {
			break
		}
	}

	return nil
}
