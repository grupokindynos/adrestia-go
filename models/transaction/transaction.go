package transaction

import (
	"github.com/grupokindynos/common/coin-factory/coins"
)

// Placeholder transaction
type PTx struct {
	ToCoin   string
	FromCoin string
	Amount   float64
	BtcRate  float64
}

type ExchangeSell struct {
	FromCoin coins.Coin
	ToCoin   coins.Coin
	Amount   float64
}
