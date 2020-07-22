package services

import (
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/plutus"
)

type HestiaService interface {
	GetAdrestiaCoins() (availableCoins []hestia.Coin, err error)
	GetExchanges() ([]hestia.ExchangeInfo, error)
	GetWithdrawals(includeComplete bool, sinceTimestamp int64, balancerId string) ([]hestia.SimpleTx, error)
	GetBalanceOrders(includeComplete bool, sinceTimestamp int64) ([]hestia.BalancerOrder, error)
	GetBalancer() (hestia.Balancer, error)
	GetDeposits(includeComplete bool, sinceTimestamp int64) ([]hestia.SimpleTx, error)
	CreateDeposit(simpleTx hestia.SimpleTx) (string, error)
	CreateWithdrawal(simpleTx hestia.SimpleTx) (string, error)
	CreateBalancerOrder(balancerOrder hestia.BalancerOrder) (string, error)
	CreateBalancer(balancer hestia.Balancer) (string, error)
	UpdateDeposit(simpleTx hestia.SimpleTx) (string, error)
	UpdateBalancer(balancer hestia.Balancer) (string, error)
	UpdateWithdrawal(simpleTx hestia.SimpleTx) (string, error)
	UpdateBalancerOrder(order hestia.BalancerOrder) (string, error)
	// Bitcou Payment
	GetVouchersByStatusV2(status hestia.VoucherStatusV2) ([]hestia.VoucherV2, error)
}

type PlutusService interface {
	GetWalletBalance(ticker string) (balance plutus.Balance, err error)
	WithdrawToAddress(body plutus.SendAddressBodyReq) (txId string, err error)
	GetAddress(coin string) (string, error)
}
