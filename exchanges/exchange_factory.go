package exchanges

import (
	"errors"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"log"
	"strings"

	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/joho/godotenv"
)

type ExchangeFactory struct {
	Hestia services.HestiaService
	Obol   obol.ObolService
	exchangesInfo map[string]hestia.ExchangeInfo
}

func NewExchangeFactory(obol obol.ObolService, hestiaDb services.HestiaService) *ExchangeFactory {
	if err := godotenv.Load(); err != nil {
		log.Println("Not .env file found for ExchangeFactory")
	}

	exFactory := new(ExchangeFactory)
	exchanges, err := hestiaDb.GetExchanges()
	if err != nil {
		log.Println("Couldn't get exchanges info", err)
	}

	exFactory.exchangesInfo = make(map[string]hestia.ExchangeInfo)
	for _, exchange := range exchanges {
		exFactory.exchangesInfo[exchange.Name] = exchange
	}

	exFactory.Obol = obol
	exFactory.Hestia = hestiaDb

	return exFactory
}

func (e *ExchangeFactory) createInstance(name string) (Exchange, error) {
	var exName = strings.ToLower(name)
	if exName == "binance" {
		return NewBinance(e.exchangesInfo[exName]), nil
	} else if exName == "southxchange" {
		return NewSouthXchange(e.exchangesInfo[exName]), nil
	} else if exName == "stex" {
		return NewStex(e.exchangesInfo[exName])
	} else if exName == "bittrex" {
		return NewBittrex(e.exchangesInfo[exName])
	} else if exName == "crex24" {
		return NewCrex24(e.exchangesInfo[exName]), nil
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
