package main

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/api/services"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
}

// For all implemented coins, tests that an exchange is provided
// and that an address can be retrieved from them
func TestAddresses(t *testing.T) {
	var coins = coinfactory.Coins
	log.Println("Coins to test: ", coins)
	var exchangeFactory = new(services.ExchangeFactory)

	for _, coin := range coins {
		log.Println(fmt.Sprintf("Getting Address for %s", coin.Name))
		ex, err := exchangeFactory.GetExchangeByCoin(*coin)
		// assert.NotNil(t, ex) // TODO Uncomment when all exchanges are implemented
		if err != nil {
			s := fmt.Sprintf("exchange not implemented for %s", coin.Name)
			color.Warn.Tips(s)
		} else {
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

// Makes sure OneConversion to BTC is implemented for every coin
func TestRateToBtc(t *testing.T) {
	var coins = coinfactory.Coins
	var exchangeFactory = new(services.ExchangeFactory)

	for _, coin := range coins {
		log.Println(fmt.Sprintf("Getting Rates for %s", coin.Name))
		ex, err := exchangeFactory.GetExchangeByCoin(*coin)
		// assert.NotNil(t, ex) // TODO Uncomment when all exchanges are implemented

		if err != nil {
			s := fmt.Sprintf("exchange not implemented for %s", coin.Name)
			color.Warn.Tips(s)
		} else {
			rate, _ := ex.OneCoinToBtc(*coin)
			assert.Greater(t, rate, 0.0)
		}
	}
}

func TestBalances(t *testing.T) {
	exchangesToTest := [...]string{"binance", "cryptobridge", "bitso"}
	var exFactory = new(services.ExchangeFactory)

	for _, exName := range exchangesToTest {
		s := fmt.Sprintf("Retrieving Balances for %s", exName)
		color.Info.Tips(s)
		ex, err := exFactory.GetExchangeByName(exName)
		if err != nil {
			s = fmt.Sprintf("exchange %s not implemented", exName)
			color.Warn.Tips(s)
		} else {
			balances, _ := ex.GetBalances()
			assert.NotNil(t, balances)
		}

	}
}