package exchanges

import (
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/hestia"
)

type IExchange interface {
	GetName() (string, error)
	GetAddress(coin coins.Coin) (string, error)
	OneCoinToBtc(coin coins.Coin) (float64, error)
	GetBalances() ([]balance.Balance, error)
	SellAtMarketPrice(SellOrder transaction.ExchangeSell) (bool, error)
	Withdraw(coin coins.Coin, address string, amount float64) (bool, error)
	GetRateByAmount(sell transaction.ExchangeSell) (float64, error)
	GetOrderStatus(orderId string) (hestia.AdrestiaStatus, error)
}

type Exchange struct {
	IExchange
	Name 	string
	BaseUrl string
}