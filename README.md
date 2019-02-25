# Skeeter

In its current form, `skeeter` is a CLI program to bridge the TPlink HS220 with MQTT.

This is very much alpha code, but I'm currently using it now to integrate my HS220 switches
to both Home Assistant and Node-Red.

The code is structured to make it (somewhat) easy to add support for other devices.  Modules
exist under the `internal/pkgs/interface` directory.  For TP-Link:
```
├── internal
│   └── pkg
│       ├── interfaces
│       │   ├── interfaces.go
│       │   ├── tplink
│       │   │   ├── interface.go
│       │   │   ├── kasa.go
│       │   │   └── types.go
│       │   └── tplink.go
```

## Usage

Devices are defined in a json file, for the moment.  This isn't the best format, and it
may change in future versions.

Example configuration, for 2 tp-link HS-220 devices:

```
{
    "modules": [
        {
            "class": "tplink",
            "device": [
                { "ip": "192.168.1.176",
                  "port": "9999",
                  "id": "b0:be:76:a9:e0:0a",
                  "type": "switch",
                  "sub-suffixes": [ "set-state",  "set-brightness" ],
                  "pub-suffixes": [ "state",  "brightness" ],
                  "poll-interval": 500
                },
                { "ip": "192.168.1.177",
                  "port": "9999",
                  "id": "b0:be:76:a9:e0:0d",
                  "type": "switch",
                  "sub-suffixes": [ "set-state",  "set-brightness" ],
                  "pub-suffixes": [ "state",  "brightness" ],
                  "poll-interval": 500
                }                
            ]
        }
    ]
}
```

MQTT topics to listen/publish to are derived from class, id, type, and sub/pub suffixes.

For the tplink hs220, right now two subscription and publish topics are supported:
relay state, and brightness level.  The must be specified in the json file in that
order.

For the above example, the first skeeter will listen to the following topics:

*subscriptions*:
```
tplink/switch/b0:be:76:a9:e0:0a/set-state
tplink/switch/b0:be:76:a9:e0:0a/set-brightness
```

*publish*:
```
tplink/switch/b0:be:76:a9:e0:0a/state
tplink/switch/b0:be:76:a9:e0:0a/brightness
```

When skeeter sees a subscription that matches, it is passsed to the tplink
module which then sets the corresponding value on the HS220.

### Command-line options
```
Usage of ./skeeter:
  -config string
        configuration file (default "/home/user/.skeeter.json")
  -password string
        MQTT broker password
  -qos int
        MQTT QoS value (default 2)
  -retained
        MQTT broker retains last message (default true)
  -server string
        MQTT broker (default "tcp://192.168.15.2:1883")
  -username string
        MQTT broker username
  -verbosity string
        Verbosity level: debug, info, errors (default "errors")
```

### Example integration with Home Assistant

From my `configuration.yaml` file:
```
  - platform: mqtt
    name: "Upstairs Recessed Lights"
    state_topic: "tplink/switch/b0:be:76:a9:e0:01/state"
    command_topic: "tplink/switch/b0:be:76:a9:e0:01/set-state"
    brightness_state_topic: "tplink/switch/b0:be:76:a9:e0:01/brightness"
    brightness_command_topic: "tplink/switch/b0:be:76:a9:e0:01/set-brightness"
    brightness_scale: 100
    payload_on: "1"
    payload_off: "0"
    retain: true
    qos: 2
```
