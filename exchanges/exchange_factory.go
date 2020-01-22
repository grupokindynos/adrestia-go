package exchanges

import (
	"errors"
	"strings"

	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/joho/godotenv"
)

type ExchangeFactory struct {
	exchangesMp map[string]IExchange
}

func NewExchangeFactory(params Params) *ExchangeFactory {
	if err := godotenv.Load(); err != nil {
		panic("you need .env at the root of adrestia-go/")
	}

	BinanceInstance := NewBinance(params)
	BitsoInstance := NewBitso(params)
	SouthInstance := NewSouthXchange(params)

	exFactory := new(ExchangeFactory)

	exFactory.exchangesMp = map[string]IExchange{
		"binance":      BinanceInstance,
		"bitso":        BitsoInstance,
		"southxchange": SouthInstance,
	}

	return exFactory
}

func (e *ExchangeFactory) GetExchangeByCoin(coin coins.Coin) (IExchange, error) {
	coinInfo, _ := coinfactory.GetCoin(coin.Info.Tag)
	exchange, ok := e.exchangesMp[coinInfo.Rates.Exchange]
	if !ok {
		return nil, errors.New("exchange not found for " + coin.Info.Tag)
	}
	return exchange, nil
}

func (e *ExchangeFactory) GetExchangeByName(name string) (IExchange, error) {
	var exName = strings.ToLower(name)
	exchange, ok := e.exchangesMp[exName]
	if !ok {
		return nil, errors.New("exchange" + name + " not found")
	}
	return exchange, nil
}
