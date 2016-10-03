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
