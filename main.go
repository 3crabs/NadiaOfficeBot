package main

import (
	"NadiaOfficeBot/db"
	"NadiaOfficeBot/files"
	"errors"
	"fmt"
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
var isWaterArabic = false

var badUsers = map[string]bool{}

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
			if t.Day()%10 < 3 && files.ReadFikus() == false {
				files.SaveFikus(true)
				_, _ = bot.Send(tgbot.NewMessage(chatId, "Пришло время полить фикус)"))
			}
			// в понедельник, среду и пятницу
			if (int(t.Weekday()) == 1 || int(t.Weekday()) == 3 || int(t.Weekday()) == 5) && isWaterArabic == false {
				isWaterArabic = true
				_, _ = bot.Send(tgbot.NewMessage(chatId, "Пришло время полить арабику и не фикус)"))
			}
		}
	}
	// в полночь
	if t.Hour() == 0 {
		db.IsSelectedDinner = false
		isWaterFlowers = false
		isWaterArabic = false
		if t.Day()%10 == 4 {
			files.SaveFikus(false)
		}
		for login := range badUsers {
			badUsers[login] = false
		}
	}
}

var bot *tgbot.BotAPI
var chatId int64 = 0

var BadDinner = "BAD_DINNER"

//help - помощь
//dinner  - место для обеда
//flowers - о цветах
func main() {
	chatId = files.ReadChatId()
	files.ReadFikus()
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

		// обработчик кнопок
		if update.CallbackQuery != nil {
			who := update.CallbackQuery.From.UserName
			what := update.CallbackQuery.Data

			// плохой выбор обеда
			if what == BadDinner {
				if badUsers[who] == false {
					db.IsSelectedDinner = false
					badUsers[who] = true
					selectDinner(fmt.Sprintf("Ну если @%s против. ", who))
				} else {
					msg := tgbot.NewMessage(chatId, fmt.Sprintf("@%s, ты сегодня уже осуждал.", who))
					_, _ = bot.Send(msg)
				}
			}
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
			selectDinner("")
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
					"- О поливе фикуса в 10:00 примерно каждый 10 день (в выходные надо отдыхать).\n"+
					"- А опрыскивать цветы надо каждый рабочий день, в 18:00 ждите напоминание."))
			continue
		}

		msg := tgbot.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		_, _ = bot.Send(msg)
	}
}

func selectDinner(prefix string) {
	if db.IsSelectedDinner {
		prefix += "Я уже выбрала. "
	}
	msg := tgbot.NewMessage(chatId, prefix+"Предлагаю сходить сегодня в '"+db.GetRandomDinnerPlace()+"'")
	msg.ReplyMarkup = tgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbot.InlineKeyboardButton{
			{tgbot.InlineKeyboardButton{Text: "осуждаю", CallbackData: &BadDinner}},
		},
	}
	_, _ = bot.Send(msg)
}
