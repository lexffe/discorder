package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	AuthToken string `json:"authToken"` // Not used currently, but planned

	LastChannel       string   `json:"lastChannel"`
	ListeningChannels []string `json:"listeningChannels"`
	AllPrivateMode    bool     `json:"allPrivateMode"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var decoded Config
	err = json.Unmarshal(file, &decoded)
	return &decoded, err
}

func (c *Config) Save(path string) error {

	encoded, err := json.MarshalIndent(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, encoded, os.FileMode(0755))
}
