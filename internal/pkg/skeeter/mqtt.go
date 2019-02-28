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
package skeeter

import (
	"crypto/tls"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
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
	log.Debugf("MQTT rx: %s: %s\n", msg.Topic(), msg.Payload())
}

func (m *MQTTOpts) ConnectionLost(client MQTT.Client, err error) {
	log.Warnf("Connection Lost!\n")
	log.Warn(err)
}

//========================================================================
// Initialize and connect to the MQTT server
//
// Username/Password and TLS are optional
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

	tlsConfig := &tls.Config{InsecureSkipVerify: true,
		ClientAuth: tls.NoClientCert}
	m.Options.SetTLSConfig(tlsConfig)

	m.Client = MQTT.NewClient(m.Options)

	token := m.Client.Connect()
	token.Wait()
	if token.Error() != nil {
		// TODO error handling for retry?
		panic(token.Error())
	}
}

//========================================================================
// Allow a module to subscribe to a topic.  We should already be connected
// to the broker before calling.
func (m *MQTTOpts) AddSubscription(topic string, handler MQTT.MessageHandler) {

	if m.Options == nil {
		log.Errorf("MQTT has not been initialized\n")
		return
	}

	log.Infof("MQTT: Adding subscription for %s\n", topic)
	token := m.Client.Subscribe(topic, m.Qos, handler)
	token.Wait()
	if token.Error() != nil {
		//TODO error handling
		panic(token.Error())
	}
}

//========================================================================
func (m *MQTTOpts) Publish(topic string, payload string) {
	log.Debugf("MQTT: publish: %s (%s)\n", topic, payload)
	token := m.Client.Publish(topic, m.Qos, m.Retained, payload)
	token.Wait()
	if token.Error() != nil {
	    log.Errorf("MQTT Publish failed: '%s'\n", token.Error())
	}
}
