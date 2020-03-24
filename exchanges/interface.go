package exchanges

import (
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/common/hestia"
)

type Exchange interface {
	GetAddress(asset string) (string, error)
	GetBalance(asset string) (float64, error)
	SellAtMarketPrice()
	Withdraw(asset string, address string, amount float64) (string, error)
	GetOrderStatus()
	GetPair(fromCoin string, toCoin string) (models.TradeInfo, error)
	GetWithdrawalTxHash(txId string, asset string) (string, error)
	GetDepositStatus(txId string) (hestia.ExchangeOrderInfo, error)
}
