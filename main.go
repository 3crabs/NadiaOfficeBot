package main

import (
	"errors"
	"fmt"
	"github.com/FedorovVladimir/go-log/logs"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/prprprus/scheduler"
	"github.com/umputun/go-flags"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
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
var isWaterArabica = false

func task() {
	loc, _ := time.LoadLocation("Asia/Barnaul")
	t := time.Now().In(loc)
	// рабочий день
	if int(t.Weekday()) > 0 && int(t.Weekday()) < 6 {
		// в 18 часов
		if t.Hour() == 18 && isWaterFlowers == false {
			isWaterFlowers = true
			_, _ = bot.Send(tgbot.NewMessage(chatId, "Пришло время опрыскивать цветы)"))
		}
		// в 10 часов, каждый 10 день, но не в выходной
		if t.Hour() == 10 && t.Day()%10 < 3 && isWaterFikus == false {
			isWaterFikus = true
			_, _ = bot.Send(tgbot.NewMessage(chatId, "Пришло время полить фикус)"))
		}
		// в 10 часов, каждый 3 день, но не в выходной
		if t.Hour() == 10 && isWaterArabica == false {
			isWaterArabica = true
			_, _ = bot.Send(tgbot.NewMessage(chatId, "Пришло время полить арабику и не фикус)"))
		}
	}
	if t.Hour() == 0 {
		isWaterFlowers = false
	}
	if t.Day()%10 == 4 {
		isWaterFikus = false
	}
	if t.Day()%3 == 0 {
		isWaterArabica = false
	}
}

var bot *tgbot.BotAPI
var chatId int64 = 0

//help - помощь
//dinner  - место для обеда
func main() {
	readChatId()
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
			saveChatId(chatId)
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
					"\nА ещё я:\n"+
					"- напоминаю про полив цветов\n"+
					"\nНу а больше я пока ничего не умею"))
			continue
		}

		// command /dinner
		if update.Message.Text == "/dinner" {
			_, _ = bot.Send(tgbot.NewMessage(update.Message.Chat.ID,
				"Предлагаю сходить сегодня в '"+getRandomDinnerPlace()+"'"))
			continue
		}

		msg := tgbot.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		_, _ = bot.Send(msg)
	}
}

func readChatId() {
	file, _ := os.Open("chat_id.txt")
	defer file.Close()
	b, _ := ioutil.ReadAll(file)
	i, _ := strconv.Atoi(string(b))
	chatId = int64(i)
}

func saveChatId(id int64) {
	f, _ := os.Create("chat_id.txt")
	defer f.Close()
	_, _ = f.WriteString(fmt.Sprintf("%v", id))
}

func getRandomDinnerPlace() string {
	places := []string{
		"Сковородовна",
		"Мантоварка",
		"Вьетнамка",
		"Столовая",
		"Гриль №1",
		"Узбечка",
		"КФС",
	}
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(places))
	return places[n]
}
