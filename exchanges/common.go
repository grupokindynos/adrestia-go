package exchanges

import (
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/obol"
)

type Params struct {
	Obol            obol.ObolService
	Plutus          services.PlutusService
	Hestia          services.HestiaService
	ExchangeFactory IExchangeFactory
}

type OrderSide struct {
	Book             string
	Type             string
	ReceivedCurrency string
	SoldCurrency     string
}

type WithdrawConfig struct {
	PercentageFee float64
	MinimumAmount float64
	Precision     float64
}

func (wf *WithdrawConfig) GetWithdrawFeeAmount(amount float64) float64 {
	amount *= (wf.PercentageFee / 100.0)
	if amount < wf.MinimumAmount {
		return wf.MinimumAmount
	}
	return amount
}
