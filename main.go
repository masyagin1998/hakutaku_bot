// Copyright (C) 2018 Mikhail Masyagin
package main

import (
	"hakutaku_bot/libhakutaku/bot"
	"hakutaku_bot/libhakutaku/config-parser"
	"hakutaku_bot/libhakutaku/log"

	"os"
)

func main() {
	// Reading config.json.
	config, err := configParser.ReadConfig()
	if err != nil {
		os.Exit(1)
	}

	// Creating logger.
	logger, err := log.NewLogger(config.LogFile)
	if err != nil {
		os.Exit(1)
	}
	defer logger.CloseLogger()

	// Running bot.
	if err = logger.Error("Error occured, while running bot", "error", bot.StartHakutakuBot(config, logger)); err != nil {
		os.Exit(1)
	}
}
