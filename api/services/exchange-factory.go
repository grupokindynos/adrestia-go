package services

import (
	"errors"
	"github.com/grupokindynos/adrestia-go/api/exchanges"
	"github.com/grupokindynos/common/coin-factory/coins"
	"strings"
)

type ExchangeFactory struct {
	Exchanges []exchanges.IExchange
}

func (e ExchangeFactory)GetExchangeByCoin(coin coins.Coin) (exchanges.IExchange, error) {
	var coinName = strings.ToLower(coin.Tag)
	if coinName == "polis" || coinName == "xsg" || coinName == "colx"{
		return exchanges.NewCryptobridge(), nil
	}
	if coinName == "btc" || coinName == "dash" || coinName == "ltc" || coinName == "xsg" || coinName == "grs" || coinName == "xzc"{
		return exchanges.NewBinance(), nil
	}
	/*if coinName == "mnp" || coinName == "onion" || coinName == "colx"{
		return exchanges.NewCrex()
	}*/
	return *new(exchanges.Exchange), errors.New("exchange not found")
}
