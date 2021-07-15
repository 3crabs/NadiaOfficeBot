package main

import (
	"NadiaOfficeBot/db"
	"NadiaOfficeBot/files"
	"errors"
	"github.com/FedorovVladimir/go-log/logs"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/prprprus/scheduler"
	"github.com/umputun/go-flags"
	"os"
	"strings"
	"time"
)

type Opts struct {
	Token string `short:"t" long:"token" description:"Telegram api token"`
	Login string `short:"l" long:"login" description:"Telegram bot login"`
}

var opts Opts

var isWaterFlowers = false
var isWaterFikus = false

func task() {
	loc, _ := time.LoadLocation("Asia/Barnaul")
	t := time.Now().In(loc)
	// рабочий день
	if int(t.Weekday()) > 0 && int(t.Weekday()) < 6 {
		// в 18 часов
		if t.Hour() == 18 {
			// каждый день
			if isWaterFlowers == false {
				isWaterFlowers = true
				_, _ = bot.Send(tgbot.NewMessage(chatId, "Пришло время опрыскивать цветы)"))
			}
		}
		// в 10 часов
		if t.Hour() == 10 {
			// каждый 10 день
			if t.Day()%10 < 3 && isWaterFikus == false {
				isWaterFikus = true
				_, _ = bot.Send(tgbot.NewMessage(chatId, "Пришло время полить фикус)"))
			}
			// в понедельник, среду и пятницу
			if int(t.Weekday()) == 1 || int(t.Weekday()) == 3 || int(t.Weekday()) == 5 {
				_, _ = bot.Send(tgbot.NewMessage(chatId, "Пришло время полить арабику и не фикус)"))
			}
		}
	}
	// в полночь
	if t.Hour() == 0 {
		isWaterFlowers = false
		if t.Day()%10 == 4 {
			isWaterFikus = false
		}
	}
}

var bot *tgbot.BotAPI
var chatId int64 = 0

//help - помощь
//dinner  - место для обеда
//flowers - о цветах
func main() {
	chatId = files.ReadChatId()
	s, err := scheduler.NewScheduler(1000)
	if err != nil {
		panic(err)
	}
	s.Every().Second(30).Do(task)

	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	p.SubcommandsOptional = true
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			logs.LogError(errors.New("[ERROR] cli error: " + err.Error()))
		}
		os.Exit(2)
	}

	bot, err = tgbot.NewBotAPI(opts.Token)
	if err != nil {
		logs.LogError(err)
		return
	}

	u := tgbot.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {

		if chatId == 0 {
			chatId = update.Message.Chat.ID
			files.SaveChatId(chatId)
		}

		// empty message
		if update.Message == nil {
			continue
		}

		if strings.Contains(update.Message.Text, opts.Login) {
			update.Message.Text = strings.Replace(update.Message.Text, opts.Login, "", -1)
			update.Message.Text = strings.Trim(update.Message.Text, " ")
		} else {
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
					"- отправь /dinner и я предложу место для обеда\n"+
					"- отправь /flowers и я расскажу о растениях офиса\n"+
					"\nА ещё я:\n"+
					"- напоминаю про полив цветов\n"+
					"\nНу а больше я пока ничего не умею"))
			continue
		}

		// command /dinner
		if update.Message.Text == "/dinner" {
			_, _ = bot.Send(tgbot.NewMessage(update.Message.Chat.ID,
				"Предлагаю сходить сегодня в '"+db.GetRandomDinnerPlace()+"'"))
			continue
		}

		// command /flowers
		if update.Message.Text == "/flowers" {
			_, _ = bot.Send(tgbot.NewMessage(update.Message.Chat.ID,
				"В офисе три растения:\n"+
					"- фикус\n"+
					"- не фикус\n"+
					"- арабика\n\n"+
					"- О поливе арабики и не фикуса я напомню в 10:00 в понеденьник, среду и пятницу.\n"+
					"- О поливе фикуса в 10:00 примерно каждый 10 день (в выходные надо отдыхать)."))
			continue
		}

		msg := tgbot.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		_, _ = bot.Send(msg)
	}
}
