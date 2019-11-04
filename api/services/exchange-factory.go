package services

import (
	"errors"
	coinfactory "github.com/grupokindynos/common/coin-factory"
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
	"bitso":		exchanges.BitsoInstance,
	"southxchange":	exchanges.SouthInstance,
}

func (e *ExchangeFactory) GetExchangeByCoin(coin coins.Coin) (exchanges.IExchange, error) {
	// TODO Make this compatible with coinfactory
	coinInfo, _ := coinfactory.GetCoin(coin.Tag)

	exchange, ok := ex[coinInfo.Rates.Exchange]
	if !ok {
		return nil, errors.New("exchange not found for " + coin.Tag)
	}
	return exchange, nil
}

func (e *ExchangeFactory) GetExchangeByName(name string) (exchanges.IExchange, error) {
	var exName = strings.ToLower(name)
	exchange, ok := ex[exName]
	if !ok {
		return nil, errors.New("exchange" + name + " not found")
	}
	return exchange, nil
}
