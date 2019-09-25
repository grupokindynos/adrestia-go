package exchanges

import (
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/common/coin-factory/coins"
)

type ExchangeBehaviour interface {
	GetAddress(coin coins.Coin) string
	OneCoinToBtc(coin coins.Coin) float64
	GetBalances(coin coins.Coin) []balance.Balance
	SellAtMarketPrice() bool
	Withdraw(coin string, address string, amount float64) //
	GetRateByAmount(sell transaction.ExchangeSell)
}

type Exchange struct {
	ExchangeBehaviour
	Name 	string
	BaseUrl string
}