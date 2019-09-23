package services

import (
	"github.com/grupokindynos/adrestia-go/api/exchanges"
	"github.com/grupokindynos/common/coin-factory/coins"
	"strings"
)

type ExchangeFactory struct {
	Exchanges []exchanges.Exchange
}

func GetExchangeByCoin(coin coins.Coin) exchanges.Exchange{
	var coinName = strings.ToLower(coin.Tag)
	if coinName == "btc" || coinName == "dash" {
		return exchanges.NewCryptobridge().Exchange
	}
	return *new(exchanges.Exchange)
}
