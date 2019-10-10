package main

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/api/services"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestAddresses(t *testing.T) {
	var coins = coinfactory.Coins
	var exchangeFactory = new(services.ExchangeFactory)

	for _, coin := range coins {
		log.Println(fmt.Sprintf("Getting Adress for %s", coin.Name))
		var ex = exchangeFactory.GetExchangeByCoin(*coin)
		log.Println(fmt.Sprintf("Getting uses for %s", coin.Name))
		assert.NotNil(t, ex)

		address, err := ex.GetAddress(*coin)
		assert.Nil(t, err)
		assert.NotEqual(t, "", address)
	}
}