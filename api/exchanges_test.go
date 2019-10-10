package main

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/api/services"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/stretchr/testify/assert"
	"log"
	"strings"
	"testing"
)

func TestAddresses(t *testing.T) {
	var coins = coinfactory.Coins
	var exchangeFactory = new(services.ExchangeFactory)

	for _, coin := range coins {
		log.Println(fmt.Sprintf("Getting Address for %s", coin.Name))
		ex, err := exchangeFactory.GetExchangeByCoin(*coin)
		// assert.NotNil(t, ex) // TODO Uncomment when all exchanges are implemented
		if err != nil {
			exName, _ := ex.GetName()
			assert.Equal(t, strings.ToLower(exName), strings.ToLower(coin.Rates.Exchange))
			if exName == "binance" || exName == "cryptobridge" { // Implemented Exchanges
				address, err := ex.GetAddress(*coin)
				assert.Nil(t, err)
				assert.NotEqual(t, "", address)
			}
		}


	}
}