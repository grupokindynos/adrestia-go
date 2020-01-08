package exchanges

import (
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/order"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/hestia"
)

type IExchange interface {
	GetName() (string, error)
	GetAddress(coin coins.Coin) (string, error)
	OneCoinToBtc(coin coins.Coin) (float64, error)
	GetBalances() ([]balance.Balance, error)
	SellAtMarketPrice(sellOrder transaction.ExchangeSell) (bool, string, error)
	Withdraw(coin coins.Coin, address string, amount float64) (bool, error)
	GetRateByAmount(sell transaction.ExchangeSell) (float64, error)
	GetOrderStatus(order order.Order) (hestia.ExchangeStatus, error)
}

type Exchange struct {
	IExchange
	Name    string
	BaseUrl string
}
