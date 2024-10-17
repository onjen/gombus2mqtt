package main

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func createMQTTClient(config Config) (mqtt.Client, error) {
	o := mqtt.NewClientOptions()

	for _, s := range config.Servers {
		o.AddBroker(s)
	}
	o.SetClientID(config.ClientID)

	o.SetOnConnectHandler(func(client mqtt.Client) {
		fmt.Println("Connected to the broker")
	})

	client := mqtt.NewClient(o)
	return client, nil
}
