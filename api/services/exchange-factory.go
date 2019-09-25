package services

import (
	"github.com/grupokindynos/adrestia-go/api/exchanges"
	"github.com/grupokindynos/common/coin-factory/coins"
	"strings"
)

type ExchangeFactory struct {
	Exchanges []exchanges.ExchangeBehaviour
}

func (e ExchangeFactory)GetExchangeByCoin(coin coins.Coin) exchanges.ExchangeBehaviour{
	var coinName = strings.ToLower(coin.Tag)
	if coinName == "polis" || coinName == "dash" {
		return exchanges.NewCryptobridge()
	}
	return *new(exchanges.Exchange)
}
