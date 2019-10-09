package exchanges

import (
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/common/coin-factory/coins"
)

type IExchange interface {
	GetName() (string, error)
	GetAddress(coin coins.Coin) (string, error)
	OneCoinToBtc(coin coins.Coin) (float64, error)
	GetBalances() ([]balance.Balance, error)
	SellAtMarketPrice(SellOrder transaction.ExchangeSell) (bool, error)
	Withdraw(coin string, address string, amount float64) (bool, error)
	GetRateByAmount(sell transaction.ExchangeSell) (float64, error)
}

type Exchange struct {
	IExchange
	Name 	string
	BaseUrl string
}