package services

import (
	"errors"
	"github.com/grupokindynos/adrestia-go/api/exchanges"
	"github.com/grupokindynos/common/coin-factory/coins"
	"strings"
)

type ExchangeFactory struct {
	Exchanges map[string]*exchanges.IExchange
}

var Exchanges = map[string]exchanges.IExchange{
	"cryptobridge" : exchanges.CBInstance,
	"binance" : exchanges.BinanceInstance,
}

func (e ExchangeFactory)GetExchangeByCoin(coin coins.Coin) (exchanges.IExchange, error) {
	var coinName = strings.ToLower(coin.Tag)
	if coinName == "polis" || coinName == "xsg" || coinName == "colx"{
		return Exchanges["cryptobridge"], nil
	}
	if coinName == "dash" || coinName == "ltc" || coinName == "grs" || coinName == "xzc"{
		return Exchanges["binance"], nil
	}
	/*if coinName == "mnp" || coinName == "onion" || coinName == "colx"{
		return exchanges.NewCrex()
	}*/
	/*if coinName == "btc" {
		return exchanges.NewBitso()
	}*/
	return *new(exchanges.Exchange), errors.New("exchange not found")
}
