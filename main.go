package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var debugFlag = flag.Bool("d", false, "Set loglevel to debug")

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

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		go fetchAndPublish(client, config.Topic)
	}
}

func createMQTTClient(config Config) (mqtt.Client, error) {
	o := mqtt.NewClientOptions()

	for _, s := range config.Servers {
		o.AddBroker(s)
	}
	o.SetClientID(config.ClientID)

	o.SetOnConnectHandler(func(client mqtt.Client) {
		slog.Info("Connected to the broker")
	})

	client := mqtt.NewClient(o)
	return client, nil
}

func fetchAndPublish(client mqtt.Client, topic string) {
	val := fetchValue()
	client.Publish(topic, 0, false, fmt.Sprint(val))
	slog.Debug("Fetched new value and published to MQTT", "value", val)
}
