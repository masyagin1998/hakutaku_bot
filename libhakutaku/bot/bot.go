// Copyright (C) 2018 Mikhail Masyagin

/*
Package bot contains basic bot functionality.
*/
package bot

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"hakutaku_bot/libhakutaku/config-parser"
	"hakutaku_bot/libhakutaku/log"
	"hakutaku_bot/libhakutaku/mgewiki-parser"

	"github.com/go-redis/redis"
	"gopkg.in/telegram-bot-api.v4"
)

// StartHakutakuBot starts telegram bot with configuration from config.
func StartHakutakuBot(config configParser.Config, logger log.Logger) (err error) {
	// Logging configuration.
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

	// Creating Redis client.
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisNumber,
	})
	if redisClient == nil {
		err = errors.New("Can't connect to Redis")
		return
	}

	// Database initialization.
	err = MGEWiki.InitDatabase(redisClient, logger)
	if err != nil {
		return
	}

	// Run database updater.
	go MGEWiki.UpdateDatabase(config.UpdateTime, redisClient, logger)

	// Creating bot.
	var bot *tgbotapi.BotAPI
	bot, err = tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return
	}

	// Choosing should we use debug.
	bot.Debug = config.Debug

	// Choosing connection mode.
	switch config.Mode {
	case "long polling":
		return longPolling(config, redisClient, bot, logger)
	case "webhook":
		return webHook(bot, redisClient, logger)
	default:
		err = errors.New("Unknown mode")
		return
	}
}

// longPolling starts telegram bot in "long polling" mode. Better for weak server and lowload.
func longPolling(config configParser.Config, redisClient *redis.Client, bot *tgbotapi.BotAPI, logger log.Logger) (err error) {
	// Creating telegram bot.
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		return
	}

	// Running bot loop.
	logger.Info("Telegram bot started.")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if err = makeResponse(config, redisClient, bot, logger, update.Message); err != nil {
			return
		}
	}
	return
}

// randomString returns random string from array
func randomString(answers []string) string {
	rand.Seed(time.Now().UTC().UnixNano())
	return answers[rand.Intn(len(answers))]
}

// makeResponse creates response message for request message
func makeResponse(config configParser.Config, redisClient *redis.Client, bot *tgbotapi.BotAPI, logger log.Logger, requestMessage *tgbotapi.Message) (err error) {
	logger.Info("New Message", "User ID", requestMessage.From.ID,
		"User Name", requestMessage.From.FirstName+" "+requestMessage.From.LastName,
		"User Link", requestMessage.From.UserName,
		"Text", requestMessage.Text)
	// Split command and arguments.
	var command, arguments string
	if strings.Contains(requestMessage.Text, " ") {
		spaceIndex := strings.Index(requestMessage.Text, " ")
		command = requestMessage.Text[:spaceIndex]
		arguments = requestMessage.Text[(spaceIndex + 1):]
	} else {
		command = requestMessage.Text
	}

	// Choose command.
	var text, link string
	switch command {
	case "/start":
		text = randomString(config.Answers.Greetings)
	case "/help":
		text = config.Answers.Help
	case "/about":
		if arguments == "" {
			text = randomString(config.Answers.NoArguments)
		} else {
			var flag int64
			flag, err = redisClient.Exists("name:" + strings.ToLower(arguments)).Result()
			if err != nil {
				return
			}
			if flag == 0 {
				text = randomString(config.Answers.NoGirl)
			} else {
				link, err = redisClient.Get("name:" + strings.ToLower(arguments)).Result()
				if err != nil {
					return
				}
				if strings.ToLower(arguments) == "hakutaku" {
					text = randomString(config.Answers.WowAboutMe)
				}
			}
		}
	case "/list":
		if arguments == "" {
			for i := 'A'; i <= 'Z'; i++ {
				var flag int64
				flag, err = redisClient.Exists("letter:" + strings.ToUpper(string(i))).Result()
				if err != nil {
					return
				}
				if flag == 0 {
					continue
				}
				text, err = redisClient.Get("letter:" + string(i)).Result()
				text = string(i) + "\n" + text
				if err != nil {
					return
				}
				responseMessage := tgbotapi.NewMessage(requestMessage.Chat.ID, text)
				_, err = bot.Send(responseMessage)
				if err != nil {
					return err
				}
			}
			return
		}
		var flag int64
		flag, err = redisClient.Exists("letter:" + strings.ToUpper(arguments)).Result()
		if err != nil {
			return
		}
		if flag == 0 {
			text = randomString(config.Answers.NoGirlLetter)
		} else {
			text, err = redisClient.Get("letter:" + strings.ToUpper(arguments)).Result()
			text = strings.ToUpper(arguments) + "\n" + text
			if err != nil {
				return
			}
		}
	case "/bug-report":
		text = config.Support
	default:
		text = randomString(config.Answers.Doubts)
	}
	if text != "" {
		responseMessage := tgbotapi.NewMessage(requestMessage.Chat.ID, text)
		_, err = bot.Send(responseMessage)
		if err != nil {
			return err
		}
	}
	if link != "" {
		responsePhoto := tgbotapi.NewPhotoUpload(requestMessage.Chat.ID, nil)
		responsePhoto.FileID = link
		responsePhoto.UseExisting = true
		_, err = bot.Send(responsePhoto)
		if err != nil {
			return err
		}
	}
	return
}

// webHook starts telegram bot in "webHook" mode. Better for powerfull server and highload.
func webHook(bot *tgbotapi.BotAPI, redisClient *redis.Client, logger log.Logger) (err error) {
	return
}
