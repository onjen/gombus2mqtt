package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var debugFlag = flag.Bool("d", false, "Set loglevel to debug")

type Application struct {
	client mqtt.Client
	config *Config
}

func main() {
	flag.Parse()
	if *debugFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	config, err := parseConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}

	client, err := createMQTTClient(*config)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to the broker: %v", token.Error())
	}

	app := Application{
		client: client,
		config: config,
	}

	// Publish retained discovery topics

	// Publish periodic updates
	ticker := time.NewTicker(time.Duration(config.IntervalSec) * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		go app.fetchAndPublish()
	}
}

func createMQTTClient(config Config) (mqtt.Client, error) {
	o := mqtt.NewClientOptions()

	for _, s := range config.Servers {
		o.AddBroker(s)
	}
	o.SetClientID(config.ClientID)
	o.SetUsername(config.User)
	o.SetPassword(config.Password)

	o.SetOnConnectHandler(func(client mqtt.Client) {
		slog.Info("Connected to the broker")
	})

	client := mqtt.NewClient(o)
	return client, nil
}

func (app *Application) fetchAndPublish() {
	frame, err := fetchValue(app.config.Device, app.config.Address)
	if err != nil {
		slog.Error("Error fetching value", "Error", err)
		return
	}
	msg, err := json.Marshal(frame)
	if err != nil {
		slog.Error("Error marshalling frame", "Error", err)
		return
	}
	app.client.Publish(fmt.Sprintf("%v/raw", app.config.TopicPrefix), 0, false, string(msg))
	for i, v := range frame.DataRecords {
		app.client.Publish(fmt.Sprintf("%v/%v/unit", app.config.TopicPrefix, i), 0, false, v.Unit.Unit)
		app.client.Publish(fmt.Sprintf("%v/%v/value", app.config.TopicPrefix, i), 0, false, fmt.Sprintf("%f", v.Value))
	}
	slog.Debug("Fetched new value and published to MQTT", "device", frame.DeviceType, "manufacturer", frame.Manufacturer)
}
