package services

import (
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/plutus"
)

type HestiaService interface {
	GetExchanges() ([]hestia.ExchangeInfo, error)
	GetDeposits(params models.OrderParams) ([]hestia.SimpleTx, error)
	CreateDeposit(orderData hestia.SimpleTx) (string, error)
	UpdateDeposit(simpleTx hestia.SimpleTx) (string, error)
}

type PlutusService interface {
	GetWalletBalance(ticker string) (balance plutus.Balance, err error)
	WithdrawToAddress(body plutus.SendAddressBodyReq) (txId string, err error)
}
