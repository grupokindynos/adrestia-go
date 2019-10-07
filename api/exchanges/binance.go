package exchanges

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"io/ioutil"
	l "log"
	"os"
	"time"

	"github.com/rootpd/go-binance"
	"github.com/go-kit/kit/log"
	"github.com/grupokindynos/adrestia-go/api/exchanges/config"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/utils"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/obol"
)

type Binance struct {
	Exchange
	AccountName  	string
	BitSharesUrl 	string
	binanceApi   	binance.Binance
	withdrawApi		binance.Binance
}

func NewBinance() *Binance {
	c := new(Binance)
	c.Name = "Binance"
	c.BaseUrl = ""
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
	c.binanceApi = binance.NewBinance(binanceService)
	return c
}

func (b Binance) GetName() (string, error) {
	return b.Name, nil
}

func (b Binance) GetBalances(coin coins.Coin) ([]balance.Balance, error) {
	s := fmt.Sprintf("Retrieving Balances for %s", b.Name)
	l.Println(s)
	var balances []balance.Balance
	res, err := b.binanceApi.Account(binance.AccountRequest{
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})

	if err!= nil {
		return balances, err
	}

	for _, asset := range res.Balances {
		rate, _ := obol.GetCoin2CoinRates("BTC", asset.Asset)
		var b = balance.Balance{
			Ticker:     asset.Asset,
			Balance:    asset.Free,
			RateBTC:    rate,
			DiffBTC:    0,
			IsBalanced: false,
		}
		if b.Balance > 0.0 {
			balances = append(balances, b)
		}

	}
	s = utils.GetBalanceLog(balances, b.Name)
	l.Println(s)
	return balances, nil
}

func (b Binance) SellAtMarketPrice(SellOrder transaction.ExchangeSell) (bool, error) {
	panic("Not Implemented")
}

func (b Binance) Withdraw(coin string, address string, amount float64) (bool, error) {
	fmt.Printf("Retrieving Account Info for %s", b.Name)
	res, _ := b.binanceApi.Account(binance.AccountRequest{
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})
	fmt.Println("an Withdraw: ", res.CanWithdraw)

	withdrawal, err := b.binanceApi.Withdraw(binance.WithdrawRequest{
		Asset:      coin,
		Address:    address,
		Amount:     amount,
		Name:       "Adrestia-go Withdrawal",
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})
	if err != nil {
		return false, err
	}
	// TODO Binance go library has an issue signing withdrawals
	fmt.Println(withdrawal)
	fmt.Println(err)
	if withdrawal.Success {
		return withdrawal.Success, nil
	}
	return  false, errors.New("could not withdraw")

}

// TODO Missing
func (b Binance) OneCoinToBtc(coin coins.Coin) (float64, error) {
	if coin.Tag == "BTC" {
		return 1.0, nil
	}
	res, err := b.binanceApi.Ticker24(binance.TickerRequest{Symbol:coin.Tag+"BTC"})
	if err != nil {
		return 0.0, err
	}
	fmt.Println(res.LastPrice, " ", res.Volume)
	return 0.0, nil
}

func GetSettings() config.BinanceAuth {
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
