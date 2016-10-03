package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/x-cray/logrus-prefixed-formatter"

	"github.com/redBorder/rbforwarder"
	"github.com/redBorder/rbforwarder/components/httpsender"
)

const (
	defaultQueueSize = 10000
	defaultWorkers   = 1
	defaultRetries   = 0
	defaultBackoff   = 2
)

// Logger is the main logger object
var Logger = logrus.New()
var logger *logrus.Entry

var (
	configFilename *string
	debug          *bool
	version        string
)

func init() {
	configFilename = flag.String("config", "", "Config file")
	debug = flag.Bool("debug", false, "Show debug info")
	versionFlag := flag.Bool("version", false, "Print version info")

	flag.Parse()

	if *versionFlag {
		displayVersion()
		os.Exit(0)
	}

	if len(*configFilename) == 0 {
		fmt.Println("No config file provided")
		flag.Usage()
		os.Exit(0)
	}

	Logger.Formatter = new(prefixed.TextFormatter)

	// Show debug info if required
	if *debug {
		Logger.Level = logrus.DebugLevel
	}

	if *debug {
		go func() {
			Logger.Debugln(http.ListenAndServe("localhost:6060", nil))
		}()
	}
}

func main() {
	var components []interface{}

	// Initialize logger
	logger = Logger.WithFields(logrus.Fields{
		"prefix": "m2http",
	})

	// Initialize rbforwarder and components
	f := rbforwarder.NewRBForwarder(loadForwarderConfig())
	components = append(components, &httpsender.HTTPSender{Config: loadHTTPConfig()})
	f.PushComponents(components)

	// Initialize MQTT handler
	mqttHandler := NewMQTTHandler(loadMQTTConfig(),
		func(client MQTT.Client, msg MQTT.Message) {
			opts := map[string]interface{}{
				"http_endpoint": msg.Topic(),
			}

			f.Produce(msg.Payload(), opts, nil)
		},
	)

	// Process reports
	go func() {
		for r := range f.GetReports() {
			report := r.(rbforwarder.Report)

			if report.Code != 0 {
				logger.Errorln(report.Status)
			}
		}
	}()

	// Wait for ctrl-c to close the consumer
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)

	// Start getting messages
	f.Run()
	if err := mqttHandler.Run(); err != nil {
		logger.Fatal(err)
	}

	<-ctrlc
	mqttHandler.Close()
}

func displayVersion() {
	fmt.Println("M2HTTP VERSION:\t\t", version)
	fmt.Println("RBFORWARDER VERSION:\t", rbforwarder.Version)
}
