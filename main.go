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

type Device struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
}

type DiscoverPayload struct {
	Device        Device `json:"device"`
	DeviceClass   string `json:"device_class"`
	Name          string `json:"name"`
	StateTopic    string `json:"state_topic"`
	UniqueID      string `json:"unique_id"`
	Unit          string `json:"unit_of_measurement"`
	ValueTemplate string `json:"value_template"`
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

	if *debugFlag {
		app.printRawFrame()
	}

	// Publish retained discovery topics
	if config.Autodiscover.Enabled {
		app.publishAutodiscover()
	}

	// TODO set will messages for availability

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

func (app *Application) printRawFrame() {
	frame, err := fetchValue(app.config.Device, app.config.Address)
	if err != nil {
		slog.Error("Error fetching value", "Error", err)
		return
	}
	content, err := json.MarshalIndent(frame, "", "  ")
	if err != nil {
		slog.Error("Error marshaling frame", "Error", err)
		return
	}
	fmt.Println(string(content))
}

func (app *Application) publishAutodiscover() {
	frame, err := fetchValue(app.config.Device, app.config.Address)
	if err != nil {
		slog.Error("Error fetching value", "Error", err)
		return
	}

	serialNumber := fmt.Sprintf("%d", frame.SerialNumber)
	d := Device{
		Name:         app.config.Name,
		Manufacturer: frame.Manufacturer,
		Model:        frame.DeviceType,
		Identifiers:  []string{serialNumber},
	}

	for i := range frame.DataRecords {
		if !app.config.Values[i].Publish {
			continue
		}
		sensorName := app.config.Values[i].Name
		payload := DiscoverPayload{
			Device:        d,
			DeviceClass:   app.config.Values[i].DeviceClass,
			Name:          sensorName,
			StateTopic:    fmt.Sprintf("%v/%v/%v/state", app.config.TopicPrefix, d.Name, sensorName),
			UniqueID:      fmt.Sprintf("%s_%s", sensorName, serialNumber),
			Unit:          app.config.Values[i].Unit,
			ValueTemplate: "{{ value_json }}",
		}
		msg, err := json.Marshal(payload)
		if err != nil {
			slog.Error("Error marshaling config payload", "Error", err)
			return
		}
		app.client.Publish(fmt.Sprintf("homeassistant/sensor/%s/config", payload.UniqueID), 0, true, string(msg))
	}
	slog.Info("Published Home Assistant autodiscover messages")

}

func (app *Application) fetchAndPublish() {
	frame, err := fetchValue(app.config.Device, app.config.Address)
	if err != nil {
		slog.Error("Error fetching value", "Error", err)
		return
	}
	for i, v := range frame.DataRecords {
		if !app.config.Values[i].Publish {
			continue
		}
		sensorName := app.config.Values[i].Name
		app.client.Publish(fmt.Sprintf("%v/%v/%v/state", app.config.TopicPrefix, app.config.Name, sensorName), 0, false, fmt.Sprintf("%f", v.Value))
	}
	slog.Debug("Fetched new value and published to MQTT", "device", frame.DeviceType, "manufacturer", frame.Manufacturer)
}
