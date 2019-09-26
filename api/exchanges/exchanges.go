package exchanges

import (
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/common/coin-factory/coins"
)

type IExchange interface {
	GetAddress(coin coins.Coin) string
	OneCoinToBtc(coin coins.Coin) float64
	GetBalances(coin coins.Coin) []balance.Balance
	SellAtMarketPrice(SellOrder transaction.ExchangeSell) bool
	Withdraw(coin string, address string, amount float64) //
	GetRateByAmount(sell transaction.ExchangeSell)
	GetSettings()
}

type Exchange struct {
	IExchange
	Name 	string
	BaseUrl string
}