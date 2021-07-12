package main

import (
	"errors"
	"github.com/FedorovVladimir/go-log/logs"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/umputun/go-flags"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Opts struct {
	Token string `short:"t" long:"token" description:"Telegram api token"`
}

var opts Opts

func main() {
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	p.SubcommandsOptional = true
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			logs.LogError(errors.New("[ERROR] cli error: " + err.Error()))
		}
		os.Exit(2)
	}

	bot, err := tgbot.NewBotAPI(opts.Token)
	if err != nil {
		logs.LogError(err)
		return
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbot.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {

		// empty message
		if update.Message == nil {
			continue
		}

		// ping -> pong
		if strings.ToLower(update.Message.Text) == "ping" {
			_, _ = bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "pong"))
			continue
		}

		// command /start
		if update.Message.Text == "/start" {
			_, _ = bot.Send(tgbot.NewMessage(update.Message.Chat.ID, "Приве, я Надя!"))
			continue
		}

		// command /help
		if update.Message.Text == "/help" {
			_, _ = bot.Send(tgbot.NewMessage(update.Message.Chat.ID,
				"Вот чем я могу вам помочь:\n"+
					"- отправь мне ping и я отобью pong\n"+
					"\nНу а больше я пока ничего не умею"))
			continue
		}

		// command /dinner
		if update.Message.Text == "/dinner" {
			_, _ = bot.Send(tgbot.NewMessage(update.Message.Chat.ID,
				"Предлагаю сходить сегодня в '"+getRandomDinnerPlace()+"'"))
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbot.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		_, _ = bot.Send(msg)
	}
}

func getRandomDinnerPlace() string {
	places := []string{
		"Узбечка",
		"Мантоварка",
		"Вьетнамка",
		"Столовая",
		"Гриль №1",
		"КФС",
	}
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(places))
	return places[n]
}
