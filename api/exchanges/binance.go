package exchanges

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/grupokindynos/adrestia-go/api/exchanges/config"
	"io/ioutil"
	"os"
	"github.com/binance-exchange/go-binance"
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

func GetSettings() config.BinanceAuth{
	file, err := ioutil.ReadFile("api/exchanges/config/binance.json")
	if err != nil {
		panic("Could not locate settings file")
	}
	var data config.BinanceAuth
	err = json.Unmarshal([]byte(file), &data)
	fmt.Println(data)
	if err != nil {
		panic(err)
	}
	return data
}