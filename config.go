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
type Value struct {
	Publish     bool   `yaml:"publish"`
	Name        string `yaml:"name"`
	DeviceClass string `yaml:"deviceClass"`
	Unit        string `yaml:"unit"`
}

type Config struct {
	Address      int                  `yaml:"address"`
	Autodiscover HAAutodiscoverConfig `yaml:"homeassistant_autodiscover"`
	ClientID     string               `yaml:"client_id"`
	Device       string               `yaml:"device"`
	IntervalSec  int                  `yaml:"interval_sec"`
	Name         string               `yaml:"name"`
	Password     string               `yaml:"password"`
	Servers      []string             `yaml:"servers"`
	TopicPrefix  string               `yaml:"topic_prefix"`
	User         string               `yaml:"user"`
	Values       []Value              `yaml:"values"`
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

	return &c, nil
}
