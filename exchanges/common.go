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
