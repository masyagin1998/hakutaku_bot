package bot

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"hakutaku_bot/libhakutaku/config-parser"
	"hakutaku_bot/libhakutaku/mgewiki-parser"

	"github.com/mgutz/logxi/v1"
	"gopkg.in/telegram-bot-api.v4"
)

// answer struct stores bot answers.
type answer struct {
	Greetings   []string `json:"greetings"`
	Help        string   `json:"help"`
	Doubt       []string `json:"doubt"`
	NoArguments []string `json:"noArguments"`
	NoGirl      []string `json:"noGirl"`
	WowAboutMe  []string `json:"wowAboutMe"`
}

// StartHakutakuBot starts telegram bot.
func StartHakutakuBot(config configParser.Config) (err error) {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return
	}

	bot.Debug = config.Debug

	switch config.Mode {
	case "long polling":
		return longPolling(bot)
	case "webhook":
		return webHook(bot)
	default:
		err = errors.New("Unknown mode")
		return
	}
}

// longPolling starts telegram bot in "long polling" mode. Better for weak server and lowload.
func longPolling(bot *tgbotapi.BotAPI) (err error) {
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

		log.Info("New message", "UserName", update.Message.From.UserName, "command", update.Message.Text)

		err = response(update.Message, bot)
		if err != nil {
			return
		}
	}
	return
}

// webHook starts telegram bot in "webHook" mode. Better for powerfull server and highload.
func webHook(bot *tgbotapi.BotAPI) (err error) {
	return
}

// response creates response message for request message
func response(requestMessage *tgbotapi.Message, bot *tgbotapi.BotAPI) (err error) {
	var command, arguments string
	if strings.Contains(requestMessage.Text, " ") {
		spaceIndex := strings.Index(requestMessage.Text, " ")
		command = requestMessage.Text[:spaceIndex]
		arguments = requestMessage.Text[(spaceIndex + 1):]
	} else {
		command = requestMessage.Text
	}

	var text, link string
	var list []string
	switch command {
	case "/start":
		text = generateGreatingString()
	case "/help":
		text = generateHelpString()
	case "/about":
		if arguments == "" {
			text = generateNoArguments()
		} else {
			link, err = MGEWikiParser.FindGirl(arguments)
			if err != nil {
				text = generateSorryString()
			} else {
				if link == "" {
					text = generateNoGirlString()
				} else {
					if arguments == "Hakutaku" {
						text = generateWowAboutMeString()
					}
				}
			}
		}
	case "/list":
		list, err = MGEWikiParser.GetListOfGirls()
		if err != nil || len(list) == 0 {
			text = generateSorryString()
			responseMessage := tgbotapi.NewMessage(requestMessage.Chat.ID, text)
			_, err = bot.Send(responseMessage)
			if err != nil {
				return err
			}
		} else {
			for _, v := range list {
				responseMessage := tgbotapi.NewMessage(requestMessage.Chat.ID, v)
				_, err = bot.Send(responseMessage)
				if err != nil {
					return err
				}
			}
		}
	default:
		text = generateDoubtString()
	}
	if command != "list" {
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
	} else {

	}
	return
}

func random(max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max)
}

func generateGreatingString() (greeting string) {
	greetings := []string{
		"Hello! Nice to meet you",
		"Hi! What do you want to know about MGE World?",
		"Howdy! How are you?",
		"Hiya! Nice to see you",
		"Hello! What's new?",
		"Such a good day! I'am Hakutaku",
		"G'day, man!",
		"What a lovely day",
	}
	return greetings[random(len(greetings))]
}

func generateHelpString() (help string) {
	return "I'am Hakutaku and I know everything about MGE world.\n" +
		"My list of commands:\n" +
		"/start - We'll start our chat\n" +
		"/help - I'll show you list of commands\n" +
		"/about - I'll try to find info about girl\n"
}

func generateDoubtString() (doubt string) {
	doubts := []string{
		"Is this a joke?!",
		"Hmmm... I don't understand you...",
		"What did you mean?",
		"Please, say it again!",
		"I love random strings too: sdfgegedrgjswrgfjeiopgketp[hege",
		"Ahaha, what did you say?",
		"Sounds stupid...",
		"??",
		"Mmm...",
		"LoL, I think, it's stupid",
	}
	return doubts[random(len(doubts))]
}

func generateNoArguments() (noArguments string) {
	return "I'am not so stupid))0)))"
}

func generateSorryString() (sorry string) {
	return "Sorry (("
}

func generateNoGirlString() (noGirl string) {
	return "There is no such a girl"
}

func generateWowAboutMeString() (wowAboutMe string) {
	return "Wow, it's about me"
}
