package services

import (
	"github.com/grupokindynos/adrestia-go/models/adrestia"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/plutus"
)

type HestiaService interface {
	GetCoinConfiguration() ([]hestia.Coin, error)
	GetBalancingOrders() ([]hestia.AdrestiaOrder, error)
	CreateAdrestiaOrder(orderData hestia.AdrestiaOrder) (string, error)
	UpdateAdrestiaOrder(orderData hestia.AdrestiaOrder) (string, error)
	GetAllOrders(adrestiaOrderParams adrestia.OrderParams) ([]hestia.AdrestiaOrder, error)
}

type PlutusService interface {
	GetWalletBalances() []balance.Balance
	GetBtcAddress() (string, error)
	GetAddress(coin string) (string, error)
	WithdrawToAddress(body plutus.SendAddressBodyReq) (txId string, err error)
	GetWalletTx(coin string, txId string) (transaction plutus.Transaction, err error)
}
