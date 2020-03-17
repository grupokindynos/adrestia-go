package exchanges

import (
	"errors"
	"log"
	"strings"

	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/joho/godotenv"
)

type ExchangeFactory struct {
	params Params
}

func NewExchangeFactory(params Params) *ExchangeFactory {
	if err := godotenv.Load(); err != nil {
		log.Println("Not .env file found for ExchangeFactory")
	}

	exFactory := new(ExchangeFactory)
	exFactory.params = params

	return exFactory
}

func (e *ExchangeFactory) createInstance(name string) (Exchange, error) {
	var exName = strings.ToLower(name)
	if exName == "binance" {
		return NewBinance(e.params), nil
	} else if exName == "southxchange" {
		return NewSouthXchange(e.params), nil
	} else if exName == "bitrue" {
		return NewBitrue(e.params), nil
	} else if exName == "bittrex" {
		return NewBittrex(e.Params), nil
	} else if exName == "crex24" {
		return NewCrex24(e.Params), nil
	} else {
		return nil, errors.New("Exchange not found")
	}
}

func (e *ExchangeFactory) GetExchangeByCoin(coin coins.Coin) (Exchange, error) {
	coinInfo, _ := coinfactory.GetCoin(coin.Info.Tag)
	return e.createInstance(coinInfo.Rates.Exchange)
}

func (e *ExchangeFactory) GetExchangeByName(name string) (Exchange, error) {
	return e.createInstance(name)
}