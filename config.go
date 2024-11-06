package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type HAAutodiscoverConfig struct {
	Enabled bool   `yaml:"enabled"`
	Prefix  string `yaml:"prefix"`
}

// https://www.home-assistant.io/integrations/sensor/#device-class
type Field struct {
	Publish     bool   `yaml:"publish"`
	Name        string `yaml:"name"`
	DeviceClass string `yaml:"deviceClass"`
	Unit        string `yaml:"unit"`
}

type Meter struct {
	Address  int    `yaml:"address"`
	Template string `yaml:"template"`
	Name     string `yaml:"name"`
}

type Config struct {
	Autodiscover   HAAutodiscoverConfig `yaml:"homeassistant_autodiscover"`
	ClientID       string               `yaml:"client_id"`
	Device         string               `yaml:"device"`
	MeterTemplates map[string][]Field   `yaml:"meter_templates"`
	IntervalSec    int                  `yaml:"interval_sec"`
	Meters         []Meter              `yaml:"meters"`
	Password       string               `yaml:"password"`
	ReadTimeoutMS  int                  `yaml:"read_timeout_ms"`
	Servers        []string             `yaml:"servers"`
	TopicPrefix    string               `yaml:"topic_prefix"`
	User           string               `yaml:"user"`
}

func parseConfig(filename string) (*Config, error) {
	c := Config{}
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file %v", err)
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(&c)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode yaml %v", err)
	}

	if err = validateConfig(&c); err != nil {
		return nil, fmt.Errorf("Failed to validate config: %v", err)
	}

	return &c, nil
}

func validateConfig(c *Config) error {
	for _, meter := range c.Meters {
		if _, ok := c.MeterTemplates[meter.Template]; !ok {
			return fmt.Errorf("Field template '%v' for meter '%v' is not defined!", meter.Template, meter.Name)
		}
	}
	return nil
}

func (c *Config) getMeterFields(m *Meter) []Field {
	return c.MeterTemplates[m.Template]
}
