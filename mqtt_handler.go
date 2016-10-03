// Copyright (C) 2016 Eneo Tecnologia S.L.
// Diego Fernández Barrera <bigomby@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"io/ioutil"
	"strconv"

	"github.com/Sirupsen/logrus"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// MQTTHandler wraps an MQTT client
type MQTTHandler struct {
	client  MQTT.Client
	handler MQTT.MessageHandler
	logger  *logrus.Entry

	Config MQTTConfig
}

// NewMQTTHandler creates a new instance for a MQTTHandler with a given config
// and message callback
func NewMQTTHandler(config MQTTConfig, publishHandler MQTT.MessageHandler) *MQTTHandler {
	mqttHandler := &MQTTHandler{
		Config: config,
	}

	if mqttHandler.Config.Logger == nil {
		mqttHandler.logger = logrus.NewEntry(logrus.New())
		mqttHandler.logger.Logger.Out = ioutil.Discard
	} else {
		mqttHandler.logger = mqttHandler.Config.Logger
	}

	if mqttHandler.Config.Debug {
		mqttHandler.logger.Logger.Level = logrus.DebugLevel
	}

	opts := MQTT.NewClientOptions()
	opts.SetClientID(config.ClientID)
	opts.AddBroker("tcp://" + config.Broker + ":" + strconv.Itoa(config.Port))
	opts.SetDefaultPublishHandler(
		func(client MQTT.Client, msg MQTT.Message) {
			logger.Debugf("MESSAGE: %s", msg.Payload())
			publishHandler(client, msg)
		},
	)

	mqttHandler.client = MQTT.NewClient(opts)

	return mqttHandler
}

// Run starts a connection to the broker and subscribes to the given topics
func (mqttHandler *MQTTHandler) Run() error {
	logger := mqttHandler.logger

	if token := mqttHandler.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	logger.Infof("Connected to broker: %s\n", mqttHandler.Config.Broker)

	for _, topic := range mqttHandler.Config.Topics {
		if token := mqttHandler.client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
			return token.Error()
		}
		logger.Infof("Suscribed to topic: %s", topic)
	}

	return nil
}

// Close terminates the connection with the broker
func (mqttHandler *MQTTHandler) Close() {
	logger := mqttHandler.logger
	logger.Info("Disconnecting...")
	mqttHandler.client.Disconnect(250)
}
