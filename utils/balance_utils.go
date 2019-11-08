package utils

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/balance"
)

func GetBalanceLog(balances []balance.Balance, ex string) string{
	currentAssetSum := 0.0
	expectingAssetSum := 0.0
	assetCounter := 0
	for _, asset := range balances {
		currentAssetSum += asset.GetBalanceInBtc(false)
		expectingAssetSum += asset.GetBalanceInBtc(true)
		assetCounter++
	}
	s := fmt.Sprintf( "Balances for %s retrieved. Total of %.8f BTC distributed in %d assets.\nConfirmedAssets: %.8f", ex, expectingAssetSum, assetCounter, currentAssetSum)
	return s
}
