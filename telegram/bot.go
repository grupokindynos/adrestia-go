package telegram

import (
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

type TelegramBot struct {
	telegramBot tgbotapi.BotAPI
	isWorking   bool
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func NewTelegramBot() TelegramBot {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_KEY"))
	if err != nil {
		log.Println("NewTelegramBot - " + err.Error())
		return TelegramBot{isWorking: false}
	}
	tb := TelegramBot{
		telegramBot: *bot,
		isWorking:   true,
	}

	return tb
}

func (tb *TelegramBot) IsWorking() bool {
	return tb.isWorking
}

func (tb *TelegramBot) SendMessage(message string) {
	if !tb.IsWorking() {
		return
	}
	tb.telegramBot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	chatId, _ := strconv.ParseInt(os.Getenv("TELEGRAM_CHAT_ID"), 10, 64)

	tb.telegramBot.Send(tgbotapi.NewMessage(chatId, message))
}
