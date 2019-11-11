package exchanges

import (
	"fmt"
	south "github.com/bitbandi/go-southxchange"
	"github.com/grupokindynos/adrestia-go/api/exchanges/config"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/utils"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/obol"
	"log"
	"os"
	"strings"
)

type SouthXchange struct {
	Exchange
	apiKey   	string
	apiSecret  	string
	southClient	south.SouthXchange
}

var SouthInstance = NewSouthXchange()

func NewSouthXchange() *SouthXchange {
	s := new(SouthXchange)
	s.Name = "SouthXchange"
	data := s.GetSettings()
	s.apiKey = data.ApiKey
	s.apiSecret = data.ApiSecret
	s.southClient = *south.New(s.apiKey, s.apiSecret, "user-agent")
	return s
}
func (s SouthXchange) GetName() (string, error) {
	return "southxchange", nil
}

func (s SouthXchange) Withdraw(coin coins.Coin, address string, amount float64) (bool, error) {
	res, err := s.southClient.Withdraw(address, strings.ToUpper(coin.Tag), amount)
	fmt.Println(res, err)
	if err!= nil {
		return false, err
	}
	fmt.Println("South Client Response: ",res.Status)
	return true, err
}

func (s SouthXchange) OneCoinToBtc(coin coins.Coin) (float64, error) {
	if coin.Tag == "BTC" {
		return 1.0, nil
	}
	rate, err := obol.GetCoin2CoinRatesWithAmount("https://obol-rates.herokuapp.com/", "btc", coin.Tag, fmt.Sprintf("%f", 1.0))
	if err != nil {
		return 0.0, err
	}
	return rate.AveragePrice, nil
}

func (s SouthXchange) GetSettings() config.SouthXchangeAuth {
	var data config.SouthXchangeAuth
	data.ApiKey = os.Getenv("SOUTH_API_KEY")
	data.ApiSecret = os.Getenv("SOUTH_API_SECRET")
	return data
}

func (s SouthXchange) GetBalances() ([]balance.Balance, error) {
	str := fmt.Sprintf("[GetBalances] Retrieving Balances for coins at %s", s.Name)
	log.Println(str)
	var balances []balance.Balance
	res, err := s.southClient.GetBalances()

	if err != nil {
		return balances, err
	}
	for _, asset := range res {
		rate, _ := obol.GetCoin2CoinRates("https://obol-rates.herokuapp.com/", "BTC", asset.Currency)
		var b = balance.Balance{
			Ticker:     		asset.Currency,
			ConfirmedBalance:	asset.Available,
			UnconfirmedBalance: asset.Unconfirmed,
			RateBTC:    		rate,
			DiffBTC:    		0.0,
			IsBalanced: 		false,
		}
		if b.GetTotalBalance() > 0.0 {
			balances = append(balances, b)
		}

	}
	str = utils.GetBalanceLog(balances, s.Name)
	log.Println(str)
	return balances, nil
}

func (s *SouthXchange) GetAddress(coin coins.Coin) (string, error) {
	address, err := s.southClient.GetDepositAddress(strings.ToLower(coin.Name))
	str := string(address)
	str = strings.Replace(str, "\\", "", -1)
	str = strings.Replace(str, "\"", "", -1)
	str = strings.Replace(str, "/", "", -1)
	return str, err
}