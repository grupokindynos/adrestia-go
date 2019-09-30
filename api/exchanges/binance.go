package exchanges

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/grupokindynos/adrestia-go/api/exchanges/config"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/utils"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/obol"
	"io/ioutil"
	l "log"
	"os"
	"github.com/binance-exchange/go-binance"
	"time"
)

type Binance struct {
	Exchange
	AccountName string
	BitSharesUrl string
	BinanceApi binance.Binance

}

func NewBinance() *Binance {
	c := new(Binance)
	c.Name = "Binance"
	c.BaseUrl = "https://api.crypto-bridge.org/"
	data := GetSettings()
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	hmacSigner := &binance.HmacSigner{
		Key: []byte(data.PrivateApi),
	}
	ctx, _ := context.WithCancel(context.Background())
	// use second return value for cancelling request when shutting down the app

	fmt.Println("Binance Service Building...")
	binanceService := binance.NewAPIService(
		"https://www.binance.com",
		data.PublicApi,
		hmacSigner,
		logger,
		ctx,
	)
	c.BinanceApi = binance.NewBinance(binanceService)
	return c
}

func (b Binance) GetBalances(coin coins.Coin) []balance.Balance {
	fmt.Sprintf("Retrieving Balances for %s", b.Name )
	var balances []balance.Balance
	res, _ := b.BinanceApi.Account(binance.AccountRequest{
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})
	for _, asset := range res.Balances {
		rate, _ :=  obol.GetCoin2CoinRates("BTC", asset.Asset)
		var b = balance.Balance{
			Ticker:     asset.Asset,
			Balance:    asset.Free,
			RateBTC:   	rate,
			DiffBTC:    0,
			IsBalanced: false,
		}
		if b.Balance > 0.0 {
			balances = append(balances, b)
		}

	}
	s := utils.GetBalanceLog(balances, b.Name)
	l.Println(s)
	return balances
}

func GetSettings() config.BinanceAuth{
	file, err := ioutil.ReadFile("api/exchanges/config/binance.json")
	if err != nil {
		panic("Could not locate settings file")
	}
	var data config.BinanceAuth
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		panic(err)
	}
	return data
}