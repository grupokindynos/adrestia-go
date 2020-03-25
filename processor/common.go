package processor

import (
	"errors"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/blockbook"
	"github.com/grupokindynos/common/obol"
	"strconv"
)

type Params struct {
	Hestia services.HestiaService
	Plutus services.PlutusService
	Obol obol.ObolService
}

func getPlutusReceivedAmount(tx blockbook.Tx, withdrawAddress string) (float64, error) {
	for _, txVout := range tx.Vout {
		for _, address := range txVout.Addresses {
			if address == withdrawAddress {
				value, err := strconv.ParseFloat(txVout.Value, 64)
				if err != nil {
					return 0.0, err
				}
				return value, nil
			}
		}
	}
	return 0.0, errors.New("Address not found")
}