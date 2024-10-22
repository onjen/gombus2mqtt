package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Servers  []string `yaml:"servers"`
	ClientID string   `yaml:"client_id"`
	Topic    string   `yaml:"topic"`
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
