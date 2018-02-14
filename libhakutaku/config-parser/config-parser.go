package configParser

import (
	"encoding/json"
	"os"
)

// Config struct stores bot configuration.
type Config struct {
	Token string `json:"token"`
	Mode  string `json:"mode"`
	Debug bool   `json:"debug"`
}

// ReadConfig reads bot configuration from config.json.
func ReadConfig() (config Config, err error) {
	config = Config{}

	configFile, err := os.Open("config.json")
	if err != nil {
		return
	}
	defer configFile.Close()

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		return
	}
	return
}
