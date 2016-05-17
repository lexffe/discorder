package discorder

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Email     string `json:"email"`
	AuthToken string `json:"authToken"` // Not used currently, but planned

	AllPrivateMode    bool     `json:"allPrivateMode"`
	LastChannel       string   `json:"lastChannel"`
	ListeningChannels []string `json:"listeningChannels"`
}

func LoadOrCreateConfig(path string) (*Config, error) {
	file, err := ioutil.ReadFile(path)
	if err == nil {
		var decoded Config
		err = json.Unmarshal(file, &decoded)
		return &decoded, err
	}

	log.Println("Failed loading config, creating new one")
	c := &Config{}
	err = c.Save(path)
	return c, err
}

func (c *Config) Save(path string) error {
	eencoded, err := json.MarshalIndent(c, "", "	")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, eencoded, os.FileMode(0755))
}
