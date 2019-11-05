package main

import (
	"fmt"
	"github.com/grupokindynos/common/coin-factory/coins"
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


// Test Withdrawal for SouthExchange from BitBandi's repo
func TestWithdrawSouth(t *testing.T) {
	fmt.Println("Test South Withdrawal")
	var exchangeFactory = new(services.ExchangeFactory)
	ex, err := exchangeFactory.GetExchangeByCoin(coins.Polis)
	if err != nil {
		fmt.Println(err)
	}
	val, err := ex.Withdraw(coins.Polis, "PHGPU2ncaduZ7FmEyFD9wZALiY4X1w8LhS", 1.0)
	fmt.Println(val)
}

// For all implemented coins, tests that an exchange is provided
// and that an address can be retrieved from them
func TestAddresses(t *testing.T) {
	// TODO Make for all coins
	var coinsToCheck = make(map[string]*coins.Coin)
	coinsToCheck["POLIS"] = &coins.Polis
	coinsToCheck["BTC"] = &coins.Bitcoin
	coinsToCheck["DASH"] = &coins.Dash

	log.Println("Coins to test: ", coinsToCheck)
	var exchangeFactory = new(services.ExchangeFactory)

	for _, coin := range coinsToCheck {
		log.Println(fmt.Sprintf("Getting Address for %s", coin.Name))
		ex, err := exchangeFactory.GetExchangeByCoin(*coin)
		// assert.NotNil(t, ex) // TODO Uncomment when all exchanges are implemented
		if err != nil {
			s := fmt.Sprintf("exchange not implemented for %s", coin.Name)
			color.Warn.Tips(s)
		} else {
			exName, _ := ex.GetName()
			fmt.Println(coin.Name, " at ", exName)
			assert.Equal(t, strings.ToLower(exName), strings.ToLower(coin.Rates.Exchange))
			address, err := ex.GetAddress(*coin)
			fmt.Println(coin.Name, ": ", address)
			assert.Nil(t, err)
			assert.NotEqual(t, "", address)

		}
	}
}

// Makes sure OneConversion to BTC is implemented for every coin
func TestRateToBtc(t *testing.T) {
	var coins = coinfactory.Coins
	var exchangeFactory = new(services.ExchangeFactory)

	for _, coin := range coins {
		ex, err := exchangeFactory.GetExchangeByCoin(*coin)
		// assert.NotNil(t, ex) // TODO Uncomment when all exchanges are implemented
		if err != nil {
			s := fmt.Sprintf("exchange not implemented for %s", coin.Name)
			color.Warn.Tips(s)
		} else {
			rate, _ := ex.OneCoinToBtc(*coin)
			assert.GreaterOrEqual(t, rate, 0.0)
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
