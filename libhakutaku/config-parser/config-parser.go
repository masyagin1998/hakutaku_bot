// Copyright (C) 2018 Mikhail Masyagin

/*
Package configParser contains configuration structs and their reader.
*/
package configParser

import (
	"encoding/json"
	"os"
)

// Config struct contains bot configuration.
type Config struct {
	Token         string        `json:"token"`
	Support       string        `json:"support"`
	Mode          string        `json:"mode"`
	UpdateTime    string        `json:"updateTime"`
	RedisAddr     string        `json:"redisAddr"`
	RedisPassword string        `json:"redisPassword"`
	RedisNumber   int           `json:"redisNumber"`
	LogFile       string        `json:"logFile"`
	Debug         bool          `json:"debug"`
	Answers       AnswerStrings `json:"answers"`
}

// AnswerStrings struct contains bot answers.
type AnswerStrings struct {
	Greetings    []string `json:"greetings"`
	Help         string   `json:"help"`
	NoArguments  []string `json:"noArguments"`
	NoGirl       []string `json:"noGirl"`
	NoGirlLetter []string `json:"noGirlLetter"`
	WowAboutMe   []string `json:"wowAboutMe"`
	Doubts       []string `json:"doubts"`
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
