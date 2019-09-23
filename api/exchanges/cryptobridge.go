package exchanges

import (
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/common/coin-factory/coins"
)

type Cryptobridge struct {
	Exchange
}

func NewCryptobridge() *Cryptobridge {
	c := new(Cryptobridge)
	c.Name = "Cryptobridge"
	return c
}

func (c Cryptobridge) GetAddress(coin coins.Coin) string {


	return "Missing Implementation"
}

func (c Cryptobridge) OneCoinToBtc(coin coins.Coin) float64 {
	panic("Missing Implementation")
}

func (c Cryptobridge) GetBalance(coin coins.Coin) []balance.Balance {

	panic("Missing Implementation")
}