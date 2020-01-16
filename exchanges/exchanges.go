package exchanges

import (
	"github.com/grupokindynos/adrestia-go/models/balance"
	em "github.com/grupokindynos/adrestia-go/models/exchange_models"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/hestia"
)

type IExchange interface {
	GetName() (string, error)
	GetAddress(coin coins.Coin) (string, error)
	OneCoinToBtc(coin coins.Coin) (float64, error)
	GetBalances() ([]balance.Balance, error)
	GetCoinConfig(coin coins.Coin) (em.CoinConfig, error)
	SellAtMarketPrice(sellOrder transaction.ExchangeSell) (bool, string, error)
	Withdraw(coin coins.Coin, address string, amount float64) (bool, error)
	GetRateByAmount(sell transaction.ExchangeSell) (float64, error)
	GetOrderStatus(order hestia.ExchangeOrder) (hestia.ExchangeStatus, error)
	GetListingAmount(order hestia.ExchangeOrder) (float64, error)
	GetDepositStatus(txid string, asset string) (bool, error)
}

type Exchange struct {
	IExchange
	Name    string
	BaseUrl string
}
