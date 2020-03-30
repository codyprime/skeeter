# Skeeter

In its current form, `skeeter` is a CLI program to bridge the TPlink HS220
smart dimmer with MQTT.

When I first installed the HS220 in my house, it was not supported by Home
Assistant.  Work is currently underway to change that, and I expect that sooner
rather than later, HA will support the HS220 natively.  However, I realized
what I wanted was for my smart devices to (as much as feasible) all be connected
via MQTT, even if they have native support in other platforms.  That way, they
become open to most IoT software without the need for any further libraries or
added support (and they remain easily scripted on the command line via
`mosquitto_sub` and `mosquitto_pub`, without interfering with other usage).

So to that end, I started Skeeter... named tongue-in-cheek after [Mosquitto](https://github.com/eclipse/mosquitto),
the open source MQTT server.

As it is now, the code that exists scratches my initial itch, of connecting the
HS220 to MQTT.  I am using it to connect the HS220 to both my Home Assistant
and Node-Red installs.

When started, it will poll the HS220 at a specified interval, to read the relay
state and brightness values.  Upon change, the new relay state and/or brightness
is sent out as an MQTT published topic.  It also subscribes to MQTT topics, in order
to set the relay and/or brightness values on the HS2220


However, the goal of this project is a bit broader, and is structured to have
the beginnings of being able to support other devices and protocols as well.

Modules exist under the `internal/pkgs/interface` directory.  For
TP-Link:
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

Devices are defined in a yaml file.

Example configuration, for a tp-link HS-220 devices:

```yaml
---
modules:
    - class: "tplink"
      device:
          - ip: "192.168.15.176"
            port: "9999"
            id: "b0:be:76:a9:ee:0d"
            type: "switch"
            sub-suffixes:
                - "set-state"
                - "set-brightness"
                - "set-brightness-smooth"
            pub-suffixes:
                - "state"
                - "brightness"
            poll-interval: 500
```

MQTT topics to listen/publish to are derived from class, id, type, and sub/pub
suffixes.

For the tplink hs220, right now two subscription and publish topics are
supported: relay state, and brightness level.  The must be specified in the
yaml file in that order.

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
![Example Usage](docs/skeeter.gif?raw=true)


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

## Build & Install

Build:
`go get github.com/codyprime/skeeter/cmd/skeeter`

Install to `$GOPATH/bin`:
`go install github.com/codyprime/skeeter/cmd/skeeter`
