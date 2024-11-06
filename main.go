package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var debugFlag = flag.Bool("d", false, "Set loglevel to debug")
var scanFlag = flag.Bool("s", false, "Run a scan to find devices")
var printFlag = flag.Int("p", -1, "Print the response at a given primary address")

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

	val, ok := os.LookupEnv("CONFFILE")
	if !ok {
		val = "config.yaml"
	}
	config, err := parseConfig(val)

	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}

	app := Application{
		config: config,
	}

	if *scanFlag {
		app.scan()
		return
	}

	if *printFlag >= 0 {
		app.printRawFrame(*printFlag)
		return
	}

	client, err := createMQTTClient(*config)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	app.client = client

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to the broker: %v", token.Error())
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

func (app *Application) printRawFrame(address int) {
	for _, meter := range app.config.Meters {
		frame, err := fetchValue(app.config.Device, address, time.Duration(app.config.ReadTimeoutMS))
		if err != nil {
			slog.Error("Error fetching value", "Error", err)
			return
		}
		content, err := json.MarshalIndent(frame, "", "  ")
		if err != nil {
			slog.Error("Error marshaling frame", "Error", err)
			return
		}
		fmt.Printf("Printing raw response for meter '%v' at address '%v'\n", meter.Name, meter.Address)
		fmt.Println(string(content))
	}
}

func (app *Application) publishAutodiscover() {
	for _, meter := range app.config.Meters {
		frame, err := fetchValue(app.config.Device, meter.Address, time.Duration(app.config.ReadTimeoutMS))
		if err != nil {
			slog.Error("Error fetching value", "Error", err)
			return
		}

		serialNumber := fmt.Sprintf("%d", frame.SerialNumber)
		d := Device{
			Name:         meter.Name,
			Manufacturer: frame.Manufacturer,
			Model:        frame.DeviceType,
			Identifiers:  []string{serialNumber},
		}

		fields := app.config.getMeterFields(&meter)
		for i := range frame.DataRecords {
			if !fields[i].Publish {
				continue
			}
			sensorName := fields[i].Name
			payload := DiscoverPayload{
				Device:        d,
				DeviceClass:   fields[i].DeviceClass,
				Name:          sensorName,
				StateTopic:    fmt.Sprintf("%v/%v/%v/state", app.config.TopicPrefix, d.Name, sensorName),
				UniqueID:      fmt.Sprintf("%s_%s", sensorName, serialNumber),
				Unit:          fields[i].Unit,
				ValueTemplate: "{{ value_json }}",
			}
			msg, err := json.Marshal(payload)
			if err != nil {
				slog.Error("Error marshaling config payload", "Error", err)
				return
			}
			topic := fmt.Sprintf("homeassistant/sensor/%s/config", payload.UniqueID)
			app.client.Publish(topic, 0, true, string(msg))
			slog.Debug("Publishing autodiscover message", "topic", topic, "payload", string(msg))
		}
	}
	slog.Info("Published Home Assistant autodiscover messages")

}

func (app *Application) fetchAndPublish() {
	for _, meter := range app.config.Meters {
		frame, err := fetchValue(app.config.Device, meter.Address, time.Duration(app.config.ReadTimeoutMS))
		if err != nil {
			slog.Error("Error fetching value", "Error", err)
			return
		}
		fields := app.config.getMeterFields(&meter)
		for i, v := range frame.DataRecords {
			if !fields[i].Publish {
				continue
			}
			sensorName := fields[i].Name
			topic := fmt.Sprintf("%v/%v/%v/state", app.config.TopicPrefix, meter.Name, sensorName)
			app.client.Publish(topic, 0, false, fmt.Sprintf("%f", v.Value))
			slog.Debug("Fetched new value and published to MQTT", "device", frame.DeviceType, "manufacturer", frame.Manufacturer, "topic", topic, "value", v.Value)
		}
	}
}

func (app *Application) scan() {
	slog.Info("Scanning for devices, this will take a while...")
	for i := 0; i <= 250; i++ {
		slog.Debug("Checking address", "address", i)
		frame, err := fetchValue(app.config.Device, i, time.Duration(app.config.ReadTimeoutMS))
		if err != nil {
			slog.Debug("Error checking address", "address", i, "error", err)
			continue
		}
		slog.Info("Found device", "primary_address", i, "serial_number", frame.SerialNumber, "manufacturer", frame.Manufacturer, "version", frame.Version, "device_type", frame.DeviceType)
	}
}
