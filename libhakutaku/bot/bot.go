// Copyright (C) 2018 Mikhail Masyagin

/*
Package bot contains basic bot functionality.
*/
package bot

import (
	"errors"

	"hakutaku_bot/libhakutaku/config-parser"
	"hakutaku_bot/libhakutaku/log"
	"hakutaku_bot/libhakutaku/mgewiki-parser"

	"github.com/go-redis/redis"
	"gopkg.in/telegram-bot-api.v4"
)

// StartHakutakuBot starts telegram bot with configuration from config.
func StartHakutakuBot(config configParser.Config, logger log.Logger) (err error) {
	// Creatins Redis client.
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisNumber,
	})
	if redisClient == nil {
		err = errors.New("Can't connect to Redis")
		return
	}

	// Starting MGEWiki parser.
	err = MGEWiki.ParseMGEWiki(config.UpdateTime, redisClient, logger)
	if err != nil {
		return
	}

	// Creating bot.
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return
	}

	// Choosing should we use debug.
	bot.Debug = config.Debug

	// Choosing connection mode.
	switch config.Mode {
	case "long polling":
		return longPolling(bot, redisClient, logger)
	case "webhook":
		return webHook(bot, redisClient, logger)
	default:
		err = errors.New("Unknown mode")
		return
	}
}

// longPolling starts telegram bot in "long polling" mode. Better for weak server and lowload.
func longPolling(bot *tgbotapi.BotAPI, redisClient *redis.Client, logger log.Logger) (err error) {
	// Creating telegram bot.
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		return
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		logger.Info("New message", "UserName", update.Message.From.UserName, "command", update.Message.Text)
	}
	return
}

// webHook starts telegram bot in "webHook" mode. Better for powerfull server and highload.
func webHook(bot *tgbotapi.BotAPI, redisClient *redis.Client, logger log.Logger) (err error) {
	return
}
