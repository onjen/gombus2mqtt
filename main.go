package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	config, err := parseConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}
	fmt.Printf("%+v\n", config)

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
