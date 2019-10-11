package services

import (
	"errors"
	"strings"

	"github.com/grupokindynos/adrestia-go/api/exchanges"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/joho/godotenv"
)

type ExchangeFactory struct {
	Exchanges map[string]*exchanges.IExchange
}

func init() {
	if err := godotenv.Load(); err != nil {
		panic("you need .env at the root of api/")
	}
}

var ex = map[string]exchanges.IExchange{
	"cryptobridge": exchanges.CBInstance,
	"binance":      exchanges.BinanceInstance,
}

func (e *ExchangeFactory) GetExchangeByCoin(coin coins.Coin) (exchanges.IExchange, error) {
	var coinName = strings.ToLower(coin.Tag)

	if coinName == "polis" || coinName == "colx" {
		return ex["cryptobridge"], nil
	}
	if coinName == "dash" || coinName == "ltc" || coinName == "grs" || coinName == "xzc" {
		return ex["binance"], nil
	}
	/*if coinName == "xsg" {
		return exchanges.NewStex()
	}*/
	/*if coinName == "mnp" || coinName == "onion" || coinName == "colx"{
		return exchanges.NewCrex()
	}*/
	/*if coinName == "btc" {
		return exchanges.NewBitso()
	}*/

	return *new(exchanges.Exchange), errors.New("exchange not found")
}
