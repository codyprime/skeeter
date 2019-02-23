package tplink

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
)

type PrefState struct {
	Index      int `json:"index"`
	Brightness int `json:"brightness"`
}

type NextAction struct {
	Type int `json:"type"`
}

type GetSysInfo struct {
	SwVer      string      `json:"sw_ver"`
	HwVer      string      `json:"hw_ver"`
	MicType    string      `json:"mic_type"`
	Model      string      `json:"model"`
	MAC        string      `json:"mac"`
	DevName    string      `json:"dev_name"`
	Alias      string      `json:"alias"`
	RelayState int         `json:"relay_state"`
	Brightness int         `json:"brightness"`
	OnTime     int         `json:"on_time"`
	ActiveMode string      `json:"active_mode"`
	Feature    string      `json:"feature"`
	Updating   int         `json:"updating"`
	IconHash   string      `json:"icon_hash"`
	RSSI       int         `json:"rssi"`
	LEDOff     int         `json:"led_off"`
	Longitude  int         `json:"longitude_i"`
	Latitude   int         `json:"latitude_i"`
	HwID       string      `json:"hwId"`
	FwID       string      `json:"fwId"`
	DeviceID   string      `json:"deviceId"`
	OEMID      string      `json:"oemId"`
	PrefState  []PrefState `json:"preferred_state"`
	NextAction NextAction  `json:"next_action"`
	ErrCode    int         `json:"err_code"`
}

type System struct {
	GetSysInfo GetSysInfo `json:"get_sysinfo"`
}

type SystemResponse struct {
	System System `json:"system"`
}

type TxSystemMsg struct {
	TxGetSysinfo struct{} `json:"get_sysinfo"`
}

type SystemMsgGetInfo struct {
	TxSystemMsg TxSystemMsg `json:"system"`
}

type Cmds int

const (
	CMD_ERR        Cmds = -1
	CMD_INFO       Cmds = 0
	CMD_RELAY      Cmds = 1
	CMD_BRIGHTNESS Cmds = 2
)

type MsgSend struct {
	Cmd  Cmds
	Data []byte
}

type MsgResp struct {
	Cmd  Cmds
	Len  uint32
	Data []byte
}

const RX_MAX_LEN = 4096

//========================================================================
// Simple XOR encode
// See: https://github.com/softScheck/tplink-smartplug/blob/master/tplink_smartplug.py
//
// ---------------------------------------------------------------
// | uint32	  | length of message
// ---------------------------------------------------------------
// | []byte   | json command; first byte xor'ed with 0xAB.
// |          | each subsequent byte xor'ed with the previous
// ---------------------------------------------------------------
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

//========================================================================
// Simple XOR decode
func kasaDecode(enc []byte) (msg []byte) {
	msg = make([]byte, len(enc))

	key := byte(171)
	for i := 0; i < len(enc); i++ {
		msg[i] = enc[i] ^ key
		key = enc[i]
	}
	return msg
}

const DIMMER_CMD = "{\"smartlife.iot.dimmer\":"

//========================================================================
// For a given message cmd, select the correct json-message and encode it.
func (c *Cmds) Marshal(data []byte) (encCmd []byte, err error) {
	var encMsg []byte
	var msg []byte

	switch *c {
	case CMD_INFO:
		getInfo := SystemMsgGetInfo{}
		msg, err = json.Marshal(getInfo)
		if err != nil {
			panic(err)
		}
	case CMD_RELAY:
		msg = []byte(fmt.Sprintf("{\"system\": {\"set_relay_state\": {\"state\": %s}}}",
			string(data)))
	case CMD_BRIGHTNESS:
		msg = []byte(fmt.Sprintf("%s {\"set_brightness\": {\"brightness\": %s}}}",
			DIMMER_CMD, string(data)))
	}

	fmt.Printf("Sending: %s\n", msg)
	encMsg = kasaEncode(msg)
	return encMsg, nil
}

//========================================================================
// Decode the response
func (c *Cmds) Unmarshal(msg []byte) (value interface{}, err error) {

	switch *c {
	case CMD_INFO:
		resp := kasaDecode(msg[:])
		giResp := SystemResponse{}
		err = json.Unmarshal(resp, &giResp)
		value = &giResp
	}

	return value, err
}