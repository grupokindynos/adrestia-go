package utils

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/balance"
)

func GetBalanceLog(balances []balance.Balance, ex string) string{
	assetSum := 0.0
	assetCounter := 0
	for _, asset := range balances {
		assetSum+= asset.Balance * asset.RateBTC
		assetCounter++
	}
	s := fmt.Sprintf( "Balances for %s retrieved. Total of %f BTC distributed in %d assets.", ex, assetSum, assetCounter )
	return s
}
