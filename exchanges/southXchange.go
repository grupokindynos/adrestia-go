package exchanges

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	south "github.com/bitbandi/go-southxchange"
	"github.com/grupokindynos/adrestia-go/exchanges/config"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/exchange_models"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/adrestia-go/utils"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
)

type SouthXchange struct {
	Exchange
	apiKey      string
	apiSecret   string
	southClient south.SouthXchange
	Obol        obol.ObolService
}

func NewSouthXchange(params exchange_models.Params) *SouthXchange {
	s := new(SouthXchange)
	s.Name = "SouthXchange"
	data := s.getSettings()
	s.apiKey = data.ApiKey
	s.apiSecret = data.ApiSecret
	s.southClient = *south.New(s.apiKey, s.apiSecret, "user-agent")
	s.Obol = params.Obol
	return s
}

func (s *SouthXchange) GetName() (string, error) {
	return "southxchange", nil
}

func (s *SouthXchange) GetAddress(coin coins.Coin) (string, error) {
	address, err := s.southClient.GetDepositAddress(strings.ToLower(coin.Tag))
	str := string(address)
	str = strings.Replace(str, "\\", "", -1)
	str = strings.Replace(str, "\"", "", -1)
	str = strings.Replace(str, "/", "", -1)
	return str, err
}

func (s *SouthXchange) OneCoinToBtc(coin coins.Coin) (float64, error) {
	if coin.Tag == "BTC" {
		return 1.0, nil
	}
	rate, err := s.Obol.GetCoin2CoinRatesWithAmount("btc", coin.Tag, fmt.Sprintf("%f", 1.0))
	if err != nil {
		return 0.0, err
	}
	return rate.AveragePrice, nil
}

func (s *SouthXchange) GetBalances() ([]balance.Balance, error) {
	str := fmt.Sprintf("[GetBalances] Retrieving Balances for coins at %s", s.Name)
	log.Println(str)
	var balances []balance.Balance
	res, err := s.southClient.GetBalances()

	if err != nil {
		return balances, err
	}
	for _, asset := range res {
		rate, _ := s.Obol.GetCoin2CoinRates("BTC", asset.Currency)
		var b = balance.Balance{
			Ticker:             asset.Currency,
			ConfirmedBalance:   asset.Available,
			UnconfirmedBalance: asset.Unconfirmed,
			RateBTC:            rate,
			DiffBTC:            0.0,
			IsBalanced:         false,
		}
		if b.GetTotalBalance() > 0.0 {
			balances = append(balances, b)
		}

	}
	str = utils.GetBalanceLog(balances, s.Name)
	log.Println(str)
	return balances, nil
}

func (s *SouthXchange) SellAtMarketPrice(sellOrder transaction.ExchangeSell) (bool, string, error) {
	return false, "", errors.New("func not implemented")
}

func (s *SouthXchange) Withdraw(coin coins.Coin, address string, amount float64) (bool, error) {
	res, err := s.southClient.Withdraw(address, strings.ToUpper(coin.Tag), amount)
	fmt.Println(res, err)
	if err != nil {
		return false, err
	}
	fmt.Println("South Client Response: ", res.Status)
	return true, err
}

func (s *SouthXchange) GetRateByAmount(sell transaction.ExchangeSell) (float64, error) {
	return 0.0, errors.New("func not implemented")
}

func (s *SouthXchange) GetOrderStatus(orderId string) (hestia.AdrestiaStatus, error) {
	return -1, errors.New("func not implemented")
}

func (s *SouthXchange) getSettings() config.SouthXchangeAuth {
	var data config.SouthXchangeAuth
	data.ApiKey = os.Getenv("SOUTH_API_KEY")
	data.ApiSecret = os.Getenv("SOUTH_API_SECRET")
	return data
}
