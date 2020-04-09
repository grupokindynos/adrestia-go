package exchanges

import (
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/common/hestia"
)

type Exchange interface {
	GetAddress(asset string) (string, error)
	GetBalance(asset string) (float64, error)
	SellAtMarketPrice(order hestia.Trade) (string, error)
	Withdraw(asset string, address string, amount float64) (string, error)
	GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error)
	GetPair(fromCoin string, toCoin string) (models.TradeInfo, error)
	GetWithdrawalTxHash(txId string, asset string) (string, error)
	GetDepositStatus(addr string, txId string, asset string) (hestia.ExchangeOrderInfo, error)
	GetName() (string, error)
}
