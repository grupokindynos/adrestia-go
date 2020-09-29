package exchanges

import (
	"errors"
	"log"
	"strings"

	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"

	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/coin-factory/coins"
)

type ExchangeFactory struct {
	Hestia        services.HestiaService
	Obol          obol.ObolService
	exchangesInfo map[string]hestia.ExchangeInfo
}

func NewExchangeFactory(obol obol.ObolService, hestiaDb services.HestiaService) *ExchangeFactory {
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

func (e *ExchangeFactory) createInstance(name string, service hestia.ServiceAccount) (Exchange, error) {
	var exName = strings.ToLower(name)
	if val, ok := e.exchangesInfo[exName]; ok {
		params := models.ExchangeParams{
			Name: val.Name,
			Keys: val.Accounts[service],
		}
		switch exName {
		case "binance":
			return NewBinance(params), nil
		case "southxchange":
			return NewSouthXchange(params), nil
		case "stex":
			return NewStex(params)
		case "bittrex":
			return NewBittrex(params)
		case "crex24":
			return NewCrex24(params), nil
		case "bithumb":
			return NewBithumb(params), nil
		}
	}

	return nil, errors.New("Exchange not found")
}

func (e *ExchangeFactory) GetExchangeByCoin(coin coins.Coin, service hestia.ServiceAccount) (Exchange, error) {
	coinInfo, _ := coinfactory.GetCoin(coin.Info.Tag)
	return e.createInstance(coinInfo.Rates.Exchange, service)
}

func (e *ExchangeFactory) GetExchangeByName(name string, service hestia.ServiceAccount) (Exchange, error) {
	return e.createInstance(name, service)
}
