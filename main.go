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
	logger.Info("Configuration",
		"Token", config.Token,
		"Support", config.Support,
		"Mode", config.Mode,
		"Update Time", config.UpdateTime,
		"Redis Addr", config.RedisAddr,
		"Redis Password", config.RedisPassword,
		"Redis Number", config.RedisNumber,
		"Log File", config.LogFile,
		"Debug", config.Debug)

	// Running bot.
	if err = logger.Error("Error occured, while running bot", "error", bot.StartHakutakuBot(config, logger)); err != nil {
		os.Exit(1)
	}
}
