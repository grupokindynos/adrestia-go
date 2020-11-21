package ladon

import (
	"errors"
	"fmt"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/telegram"
	"log"
	"strconv"
	"strings"
)

type BitcouPaymentCommands struct {
	Obol   obol.ObolService
	ExFactory *exchanges.ExchangeFactory
	ExInfo []hestia.ExchangeInfo
	TgBot telegram.TelegramBot
	BitcouService services.BitcouV2Service
}

func (bpc *BitcouPaymentCommands) Start() {
	bpc.TgBot.Debug(true)
	updates, err := bpc.TgBot.GetUpdates()
	if err != nil {
		log.Println("botCommands::Start::GetUpdates" + err.Error())
		return
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		log.Println("USER ID: ", update.Message.From.ID)
		chatId := strconv.FormatInt(update.Message.Chat.ID, 10)
		if update.Message.IsCommand() {
			if !IsBotAuthorizedUser(update.Message.From.ID) {
				bpc.TgBot.SendMessage("You're not authorized to use the bot service.", chatId)
				continue
			}
			switch update.Message.Command() {
			case "info":
				bpc.sendInfo(chatId)
			case "balances":
				bpc.sendAllBalances(chatId)
			case "sethour":
				bpc.setPaymentHour(update.Message.Text, chatId)
			case "setamount":
				bpc.setMinimumPaymentAmount(update.Message.Text, chatId)
			case "floating":
				bpc.floatingAccountBalance(chatId)
			case "help":
				bpc.TgBot.SendMessage("List of available commands:" +
					"\n/info -> returns system variables info (payment hour and minimum payment amount)" +
					"\n/balances -> returns the balances of all the exchanges." +
					"\n/sethour v -> sets the payment hour. An integer number between 0 and 23 must be provided." +
					"\nex. \"/sethour 15\" sets the payment hour to 15 hrs (3:00 pm) CDT." +
					"\n/setamount v -> sets the minimum payment amount. A float number must be provided." +
					"\nex. \"/setamount 857.66\" sets the minimum payment amount to 857.66 USD.", chatId)
			default:
				bpc.TgBot.SendMessage(fmt.Sprintf("That is not a recognizable command.\nPlease refer to /help for more info"), chatId)
			}
		}
	}
}

func (bpc *BitcouPaymentCommands) sendAllBalances(chatId string) {
	rateBtcUsd, err := bpc.Obol.GetCoin2FIATRate("BTC", "USD")
	if err != nil {
		log.Println("bitcouPayment::GenerateWithdrawals::GetCoin2FIATRate::" + err.Error())
		return
	}

	var bal float64
	var currency string
	var totalBalanceUSD float64
	for _, exchange := range bpc.ExInfo {
		ex, err := bpc.ExFactory.GetExchangeByName(exchange.Name, hestia.VouchersAccount)
		if err != nil {
			bpc.TgBot.SendMessage("Unable to get " + exchange.Name, chatId)
			continue
		}
		if _, ok := BTCExchanges[exchange.Name]; ok {
			currency = "BTC"
			bal, err = ex.GetBalance("BTC")
		} else {
			currency = exchange.StockCurrency
			bal, err = ex.GetBalance(exchange.StockCurrency)
		}
		if err != nil {
			bpc.TgBot.SendMessage("Unable to get balance for " + exchange.Name, chatId)
			continue
		}

		if currency == "BTC" {
			totalBalanceUSD += bal * rateBtcUsd
		} else {
			totalBalanceUSD += bal
		}

		bpc.TgBot.SendMessage(fmt.Sprintf("Exchange: %s\nBalance: %.8f %s", exchange.Name, bal, currency), chatId)
	}

	bpc.TgBot.SendMessage(fmt.Sprintf("Total balance in USD: %.8f", totalBalanceUSD), chatId)
}

func (bpc *BitcouPaymentCommands) floatingAccountBalance(chatId string) {
	balance, err := bpc.BitcouService.GetFloatingAccountInfo()
	if err != nil {
		bpc.TgBot.SendMessage(fmt.Sprintf("Error Retrieving Floating Account Balance: %s", err.Error()), chatId)
	}
	bpc.TgBot.SendMessage(fmt.Sprintf("Floating Account Balance: € %.2f", balance/100), chatId)
}

func (bpc *BitcouPaymentCommands) setPaymentHour(message string, chatId string) {
	val, err := getValue(message)
	if err != nil {
		bpc.TgBot.SendMessage(err.Error(), chatId)
		return
	}
	hour, err := strconv.Atoi(val)
	if err != nil {
		bpc.TgBot.SendMessage("An integer number between 0 - 23 must be provided", chatId)
		return
	}

	if hour < 0 || hour > 23 {
		bpc.TgBot.SendMessage("An integer number between 0 - 23 must be provided", chatId)
		return
	}

	// correct the provided hour with the hour that will represent on the server.
	serverHour := (hour + serverHourDifference) % 24
	bitcouPaymentHourUTC = serverHour

	if len(val) == 1 {
		val = "0" + val
	}

	strServerHour := strconv.Itoa(serverHour)
	if len(strServerHour) == 1 {
		strServerHour = "0" + strServerHour
	}

	bpc.TgBot.SendMessage(fmt.Sprintf("Payment hour changed correctly." +
		"\nNew payment hour:" +
		"\n%s:00 CDT." +
		"\n%s:00 UTC.", val, strServerHour), chatId)
}

func (bpc *BitcouPaymentCommands) setMinimumPaymentAmount(message string, chatId string) {
	val, err := getValue(message)
	if err != nil {
		bpc.TgBot.SendMessage(err.Error(), chatId)
		return
	}

	value, err := strconv.ParseFloat(val, 64)
	if err != nil {
		bpc.TgBot.SendMessage("A float number must be provided", chatId)
		return
	}

	minimumPaymentAmount = value
	bpc.TgBot.SendMessage(fmt.Sprintf("Minimum payment amount changed correctly.\nNew payment amount: %.8f", value), chatId)
}

func (bpc *BitcouPaymentCommands) sendInfo(chatId string) {
	h := bitcouPaymentHourUTC - serverHourDifference
	if h < 0 {
		h += 24
	}
	cdtHour := strconv.Itoa(h % 24)
	utcHour := strconv.Itoa(bitcouPaymentHourUTC)

	if len(cdtHour) == 1 {
		cdtHour = "0" + cdtHour
	}
	if len(utcHour) == 1 {
		utcHour = "0" + utcHour
	}

	bpc.TgBot.SendMessage(fmt.Sprintf("Payment hour:" +
		"\n%s:00 CDT." +
		"\n%s:00 UTC." +
		"\nMinimum payment amount:" +
		"\n%.8f USD", cdtHour, utcHour, minimumPaymentAmount), chatId)
}

func getValue(message string) (string, error) {
	split := strings.Split(message, " ")
	if len(split) != 2 {
		return "", errors.New("Invalid format")
	}

	return split[1], nil
}