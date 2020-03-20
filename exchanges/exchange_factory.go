package exchanges

import (
	"errors"
	"github.com/grupokindynos/common/obol"
	"log"
	"strings"

	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/joho/godotenv"
)

type ExchangeFactory struct {
	Obol obol.ObolService
}

func NewExchangeFactory(obol obol.ObolService) *ExchangeFactory {
	if err := godotenv.Load(); err != nil {
		log.Println("Not .env file found for ExchangeFactory")
	}

	exFactory := new(ExchangeFactory)
	exFactory.Obol = obol

	return exFactory
}

func (e *ExchangeFactory) createInstance(name string) (Exchange, error) {
	var exName = strings.ToLower(name)
	if exName == "binance" {
		return NewBinance(e.Obol), nil
	} else if exName == "southxchange" {
		return NewSouthXchange(e.Obol), nil
	} else if exName == "bitrue" {
		return NewBitrue(e.Obol), nil
	} else if exName == "bittrex" {
		return NewBittrex(e.Obol), nil
	} else if exName == "crex24" {
		return NewCrex24(e.Obol), nil
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
