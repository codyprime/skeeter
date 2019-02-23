package skeeter

import (
	"crypto/tls"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type MQTTOpts struct {
	Server   string
	Clientid string
	Username string
	Password string
	Qos      byte
	Retained bool
	Options  *MQTT.ClientOptions
	Client   MQTT.Client
}

func msgReceived(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("MQTT rx: %s: %s\n", msg.Topic(), msg.Payload())
}

func (m *MQTTOpts) ConnectionLost(client MQTT.Client, err error) {
    fmt.Printf("Connection Lost!\n")
    fmt.Println(err)
}

//========================================================================
// Initialize and connect to the MQTT server
func (m *MQTTOpts) Init() {
	m.Options = MQTT.NewClientOptions()
	m.Options.AddBroker(m.Server)
	m.Options.SetClientID(m.Clientid)
	m.Options.SetCleanSession(true)
	m.Options.SetConnectionLostHandler(m.ConnectionLost)
	m.Options.SetAutoReconnect(true)

	if m.Username != "" {
		m.Options.SetUsername(m.Username)
		if m.Password != "" {
			m.Options.SetPassword(m.Password)
		}
	}

	fmt.Println(m)
	tlsConfig := &tls.Config{InsecureSkipVerify: true,
		ClientAuth: tls.NoClientCert}
	m.Options.SetTLSConfig(tlsConfig)

	m.Client = MQTT.NewClient(m.Options)

	token := m.Client.Connect()
	token.Wait()
	if token.Error() != nil {
		// TODO error handling
		panic(token.Error())
	}
}

//========================================================================
// Allow a module to subscribe to a topic
func (m *MQTTOpts) AddSubscription(topic string, handler MQTT.MessageHandler) {

	if m.Options == nil {
		fmt.Printf("MQTT has not been initialized\n")
		return
	}

	fmt.Printf("Adding subscription for %s\n", topic)
	token := m.Client.Subscribe(topic, m.Qos, handler)
	token.Wait()
	if token.Error() != nil {
		//TODO error handling
		panic(token.Error())
	}
}

func (m *MQTTOpts) Publish(topic string, payload string) {
	fmt.Printf("publish: %s (%s)\n", topic, payload)
	m.Client.Publish(topic, m.Qos, m.Retained, payload)
}
