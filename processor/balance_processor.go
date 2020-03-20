package processor

import (
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"log"
)

type Processor struct {
	Hestia services.HestiaService
}

var(
	exchangesMp map[string]bool
	exchangeFactory *exchanges.ExchangeFactory
)

func (p *Processor) Start() {
	coins, err := p.Hestia.GetAdrestiaCoins()
	if err != nil {
		log.Println("Unable to get adrestia coins")
		return
	}

	for _, coin := range coins {
		coinInfo, err := coinfactory.GetCoin(coin.Ticker)
		if err != nil {
			log.Println("Unable to get coin " + err.Error() + " coin: " + coin.Ticker)
			continue
		}
		exchange, err := exchangeFactory.GetExchangeByCoin(*coinInfo)
		if err != nil {
			log.Println(err)
			continue
		}
		balance, err := exchange.GetBalance("stable coin")
		if err != nil {
			log.Println(err)
			continue
		}
		if balance < 1000 // Defined stock, needs to think how I'm going to get this info
		{
			Refill(exchange)
		}
	}
}

func

func getExchanges() (map[exchanges.Exchange]bool, error) {
	exchanges := make(map[exchanges.Exchange]bool)


	return exchanges, nil
}
