package exchanges

import (
	"errors"
	"github.com/grupokindynos/common/blockbook"
	cf "github.com/grupokindynos/common/coin-factory"
	"math"
	"strconv"
)

func blockbookConfirmed(addr string, txId string, minConfirm int) (float64, error) {
	var blockExplorer blockbook.BlockBook
	coin, err := cf.GetCoin("ETH")
	if err != nil {
		return 0.0, errors.New("Unable to get ETH coin")
	}
	blockExplorer.Url = "https://" + coin.BlockchainInfo.ExternalSource
	addrInfo, err := blockExplorer.GetEthAddress(addr)
	if err != nil {
		return 0.0, errors.New("Error while getting eth address from blockbook " + err.Error())
	}
	receivedAmount := 0.0
	found := false
	for _, txInfo := range addrInfo.Transactions {
		if txInfo.Txid == txId && txInfo.Confirmations > 2 * minConfirm {
			found = true
			for _, tokenTxInfo := range txInfo.TokenTransfers {
				val, _ := strconv.ParseFloat(tokenTxInfo.Value, 64)
				receivedAmount += val / (math.Pow10(tokenTxInfo.Decimals) * 1.0)
			}
		}
	}
	if found {
		return receivedAmount, nil
	}

	return 0.0, errors.New("tx not found or still not confirmed")
}

func round(f float64) float64 {
	return math.Floor(f + .5)
}

func roundFixedPrecision(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return round(f*shift) / shift
}