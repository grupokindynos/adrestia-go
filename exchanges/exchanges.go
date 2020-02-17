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
	GetBalances() ([]balance.Balance, error)
	SellAtMarketPrice(sellOrder hestia.ExchangeOrder) (string, error)
	Withdraw(coin coins.Coin, address string, amount float64) (string, error)
	GetRateByAmount(sell transaction.ExchangeSell) (float64, error)
	GetOrderStatus(order hestia.ExchangeOrder) (hestia.OrderStatus, error)
	GetPair(fromCoin string, toCoin string) (OrderSide, error)
	GetWithdrawalTxHash(txId string, asset string, address string, withdrawalAmount float64) (string, error)
	GetDepositStatus(txid string, asset string) (hestia.OrderStatus, error)
}

type IExchangeFactory interface {
	GetExchangeByCoin(coin coins.Coin) (IExchange, error)
	GetExchangeByName(name string) (IExchange, error)
}

type Exchange struct {
	IExchange
	Name    string
	BaseUrl string
}
