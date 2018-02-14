package main

import (
	"hakutaku_bot/libhakutaku/bot"
	"hakutaku_bot/libhakutaku/config-parser"

	"github.com/mgutz/logxi/v1"
)

func main() {
	// Reading config.json.
	log.Info("Reading configuration file.")
	config, err := configParser.ReadConfig()
	if err != nil {
		log.Error("Error occured, while reading configuration file", "error", err)
	}

	// Starting bot.
	log.Info("Starting hakutaku_bot", "token", config.Token, "mode", config.Mode, "debug", config.Debug)
	err = bot.StartHakutakuBot(config)
	log.Error("Error occured, while running bot.", "error", err)
}
