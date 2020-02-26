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

func (e *ExchangeFactory) createInstance(name string) (IExchange, error) {
	var exName = strings.ToLower(name)
	if exName == "binance" {
		return NewBinance(e.params), nil
	} else if exName == "southxchange" {
		return NewSouthXchange(e.params), nil
	} else if exName == "bitso" {
		return NewBitso(e.params), nil
	} else {
		return nil, errors.New("exchange not found for " + exName)
	}
}

func (e *ExchangeFactory) GetExchangeByCoin(coin coins.Coin) (IExchange, error) {
	coinInfo, _ := coinfactory.GetCoin(coin.Info.Tag)
	return e.createInstance(coinInfo.Rates.Exchange)
}

func (e *ExchangeFactory) GetExchangeByName(name string) (IExchange, error) {
	return e.createInstance(name)
}
