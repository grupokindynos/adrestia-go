package processor

import (
	"errors"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/blockbook"
	cf "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/obol"
	"math"
	"strconv"
)

type Params struct {
	Hestia services.HestiaService
	Plutus services.PlutusService
	Obol obol.ObolService
}

func getPlutusReceivedAmount(addr string, txId string) (float64, error) {
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
		if txInfo.Txid == txId && txInfo.Confirmations > 0 {
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