package main

import (
	"fmt"
	"github.com/grupokindynos/common/coin-factory/coins"
	"log"
	"strings"
	"testing"

	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/api/services"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)


// TODO Make for all coins
var coinsToCheck = make(map[string]*coins.Coin)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
	coinsToCheck["POLIS"] = &coins.Polis
	coinsToCheck["BTC"] = &coins.Bitcoin
	coinsToCheck["DASH"] = &coins.Dash
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
// and that an address can be retrieved from them.
// Also requests a BTC address needed to
func TestAddresses(t *testing.T) {
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

			fmt.Println(err)
			assert.NotEqual(t, "", address)

			// Bitcoin Address
			btcAddress, err := ex.GetAddress(coins.Bitcoin)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(btcAddress)
				assert.NotEqual(t, "", btcAddress)
			}
		}
	}
}

// Makes sure OneConversion to BTC is implemented for every coin
func TestRateToBtc(t *testing.T) {
	var exchangeFactory = new(services.ExchangeFactory)

	for _, coin := range coinsToCheck {
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
	exchangesToTest := [...]string{"binance", "southxchange", "bitso"}
	var exFactory = new(services.ExchangeFactory)

	for _, exName := range exchangesToTest {
		s := fmt.Sprintf("Retrieving Balances for %s", exName)
		color.Info.Tips(s)
		ex, err := exFactory.GetExchangeByName(exName)
		if err != nil {
			s = fmt.Sprintf("exchange %s not implemented", exName)
			color.Warn.Tips(s)
		} else {
			balances, err := ex.GetBalances()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(balances)
			assert.NotNil(t, balances)
		}

	}
}

