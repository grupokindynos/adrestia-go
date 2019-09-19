package transaction

import coinfactory "github.com/grupokindynos/obol/models/coin-factory"

// Object for selling in an Exchange
type ExchangeSell struct{
	FromCoin coinfactory.Coin
	ToCoin coinfactory.Coin
	Amount float64
}
