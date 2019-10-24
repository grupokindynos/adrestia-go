package exchanges

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/apigateway"
	south "github.com/bitbandi/go-southxchange"
	"github.com/grupokindynos/adrestia-go/api/exchanges/config"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/utils"
	"github.com/grupokindynos/common/obol"
	"github.com/joho/godotenv"
	"github.com/rootpd/go-binance"
	"log"
	"os"
	"time"
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

func (s SouthXchange) GetSettings() config.SouthXchangeAuth {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
	log.Println(fmt.Sprintf("[GetSettings] Retrieving settings for Binance"))
	var data config.SouthXchangeAuth
	data.ApiKey = os.Getenv("SOUTH_API_KEY")
	data.ApiSecret = os.Getenv("BINANCE_PRIV_WITHDRAW")
	return data
}

func (so SouthXchange) GetBalances() ([]balance.Balance, error) {
	s := fmt.Sprintf("[GetBalances] Retrieving Balances for coins at %s", b.Name)
	log.Println(s)
	var balances []balance.Balance
	res, err := so.southClient.GetBalances()

	if err != nil {
		return balances, err
	}

	for _, asset := range res {
		rate, _ := obol.GetCoin2CoinRates("https://obol-rates.herokuapp.com/", "BTC", asset.Currency)
		var b = balance.Balance{
			Ticker:     asset.Currency,
			Balance:    asset.Available,
			RateBTC:    rate,
			DiffBTC:    0,
			IsBalanced: false,
		}
		if b.Balance > 0.0 {
			balances = append(balances, b)
		}

	}
	s = utils.GetBalanceLog(balances, so.Name)
	log.Println(s)
	return balances, nil
}