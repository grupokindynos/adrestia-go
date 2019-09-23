package transaction

import coinfactory "github.com/grupokindynos/common/coin-factory"

// Placeholder transaction
type PTx struct {
	ToCoin string
	FromCoin string
	Amount float64
	Rate float64
}

type ExchangeSell struct {
	FromCoin coinfactory.Coin
	ToCoin   coinfactory.Coin
	Amount   float64
}