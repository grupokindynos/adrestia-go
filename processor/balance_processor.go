package processor

import (
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/adrestia-go/services"
)

type Processor struct {
	Hestia services.HestiaService
}

func Start() {

}

func (p *Processor) getExchanges() {
	p.Hestia.GetAdrestiaCoins()
}
