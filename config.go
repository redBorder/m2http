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

	yaml "gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
	"github.com/redBorder/rbforwarder"
	"github.com/redBorder/rbforwarder/components/httpsender"
)

// MQTTConfig contains the configuration for a MQTTHandler
type MQTTConfig struct {
	ClientID string
	Broker   string
	Topics   []string
	Port     int

	Debug  bool
	Logger *logrus.Entry
}

func loadConfig(filename, component string) (config map[string]interface{}, err error) {
	generalConfig := make(map[string]interface{})
	config = make(map[string]interface{})

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	yaml.Unmarshal([]byte(data), &generalConfig)
	if err != nil {
		return
	}

	for k, v := range generalConfig[component].(map[interface{}]interface{}) {
		config[k.(string)] = v
	}

	return
}

func loadForwarderConfig() rbforwarder.Config {
	pipelineConfig, err := loadConfig(*configFilename, "pipeline")
	if err != nil {
		logger.Fatal(err)
	}

	config := rbforwarder.Config{}
	if retries, ok := pipelineConfig["retries"].(int); ok {
		config.Retries = retries
	} else {
		logger.Fatal("Invalid 'retries' option")
	}
	if backoff, ok := pipelineConfig["backoff"].(int); ok {
		config.Backoff = backoff
	} else {
		logger.Fatal("Invalid 'backoff' option")
	}
	if queue, ok := pipelineConfig["queue"].(int); ok {
		config.QueueSize = queue
	} else {
		logger.Fatal("Invalid 'queue' option")
	}

	logger.WithFields(map[string]interface{}{
		"retries": config.Retries,
		"backoff": config.Backoff,
		"queue":   config.QueueSize,
	}).Info("Forwarder config")

	return config
}

func loadHTTPConfig() httpsender.Config {
	httpConfig, err := loadConfig(*configFilename, "http")
	if err != nil {
		logger.Fatal(err)
	}

	config := httpsender.Config{}
	if workers, ok := httpConfig["workers"].(int); ok {
		config.Workers = workers
	} else {
		config.Workers = 1
	}
	if *debug {
		config.Logger = logger.WithField("prefix", "http sender")
		config.Debug = true
	}
	if url, ok := httpConfig["url"].(string); ok {
		config.URL = url
	} else {
		logger.Fatal("Invalid 'url' option")
	}

	logger.WithFields(map[string]interface{}{
		"workers": config.Workers,
		"debug":   config.Debug,
		"url":     config.URL,
	}).Info("HTTP config")

	return config
}

func loadMQTTConfig() MQTTConfig {
	mqttConfig, err := loadConfig(*configFilename, "mqtt")
	if err != nil {
		logger.Fatal(err)
	}

	config := MQTTConfig{}
	if broker, ok := mqttConfig["broker"].(string); ok {
		config.Broker = broker
	} else {
		logger.Fatal("Invalid 'broker' option")
	}
	if port, ok := mqttConfig["port"].(int); ok {
		config.Port = port
	} else {
		logger.Fatal("Invalid 'port' option")
	}
	if clientID, ok := mqttConfig["clientid"].(string); ok {
		config.ClientID = clientID
	} else {
		logger.Fatal("Invalid 'clientid' option")
	}
	if topics, ok := mqttConfig["topics"].([]interface{}); ok {
		for _, topic := range topics {
			config.Topics = append(config.Topics, topic.(string))
		}
	}
	config.Logger = logger.WithField("prefix", "mqtt")
	if *debug {
		config.Debug = true
	}

	return config
}
