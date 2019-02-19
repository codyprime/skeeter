package tplink

import (
	"bytes"
	"encoding/json"
	"encoding/binary"
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

type MsgResp struct {
	Cmd  Cmds
	Len  uint32
	Data []byte
}

const RX_MAX_LEN = 4096

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

func (c *Cmds) Marshal() (encCmd []byte, err error) {
	var encMsg []byte

	switch *c {
	case CMD_INFO:
		getInfo := SystemMsgGetInfo{}
		infoMsg, err := json.Marshal(getInfo)
		if err != nil {
			panic(err)
		}
		encMsg = kasaEncode(infoMsg)
	}

	return encMsg, nil
}

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
