package exchange_models

import (
	"github.com/grupokindynos/common/obol"
)

type Params struct {
	Obol obol.ObolService
}

type CoinConfig struct {
	PercentageDepositFee    float64
	PercentageWithdrawalFee float64
	DepositFee              float64
	WithdrawalFee           float64
	MinimumWithdrawal       float64
	MinimumDeposit          float64
}

func (cc *CoinConfig) GetFeeWithdrawalAmount(amount float64) float64 {
	feeAmount := amount * (cc.PercentageWithdrawalFee / 100.0)
	if feeAmount < cc.WithdrawalFee {
		return cc.WithdrawalFee
	}

	return feeAmount
}

func (cc *CoinConfig) GetFeeDepositAmount(amount float64) float64 {
	feeAmount := amount * (cc.PercentageDepositFee / 100.0)
	if feeAmount < cc.DepositFee {
		return cc.DepositFee
	}

	return feeAmount
}
