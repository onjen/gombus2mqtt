package main

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
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
		select {
		case <-ticker.C:
			go fetchAndPublish(client, config.Topic)
		}
	}
}

func fetchAndPublish(client mqtt.Client, topic string) {
	val := fetchValue()
	client.Publish(topic, 0, false, fmt.Sprint(val))
}
