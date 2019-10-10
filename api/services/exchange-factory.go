package services

import (
	"github.com/grupokindynos/adrestia-go/api/exchanges"
	"github.com/grupokindynos/common/coin-factory/coins"
	"strings"
)

type ExchangeFactory struct {
	Exchanges []exchanges.IExchange
}

func (e ExchangeFactory)GetExchangeByCoin(coin coins.Coin) exchanges.IExchange {
	var coinName = strings.ToLower(coin.Tag)
	if coinName == "polis" || coinName == "xsg" || coinName == "colx"{
		return exchanges.NewCryptobridge()
	}
	if coinName == "btc" || coinName == "dash" || coinName == "ltc" || coinName == "xsg" || coinName == "grs" || coinName == "xzc"{
		return exchanges.NewBinance()
	}
	return *new(exchanges.Exchange)
}
