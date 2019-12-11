package exchanges

import (
	"errors"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"strings"

	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/joho/godotenv"
)

type ExchangeFactory struct {

}

func init() {
	if err := godotenv.Load(); err != nil {
		panic("you need .env at the root of api/")
	}
}

var ex = map[string]IExchange{
	"cryptobridge": CBInstance,
	"binance":      BinanceInstance,
	"bitso":        BitsoInstance,
	"southxchange": SouthInstance,
}

func (e *ExchangeFactory) GetExchangeByCoin(coin coins.Coin) (IExchange, error) {
	coinInfo, _ := coinfactory.GetCoin(coin.Tag)
	exchange, ok := ex[coinInfo.Rates.Exchange]
	if !ok {
		return nil, errors.New("exchange not found for " + coin.Tag)
	}
	return exchange, nil
}

func (e *ExchangeFactory) GetExchangeByName(name string) (IExchange, error) {
	var exName = strings.ToLower(name)
	exchange, ok := ex[exName]
	if !ok {
		return nil, errors.New("exchange" + name + " not found")
	}
	return exchange, nil
}
