package exchanges

import "github.com/grupokindynos/common/hestia"

type Exchange interface {
	GetAddress(asset string) (string, error)
	GetBalance(asset string) (float64, error)
	SellAtMarketPrice()
	Withdraw()
	GetOrderStatus()
	GetPair()
	GetWithdrawalTxHash()
	GetDepositStatus(txId string) (hestia.ExchangeOrderInfo, error)
}
