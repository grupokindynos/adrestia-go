package main

// DON'T RUN THIS TEST
// it withdraws real funds from your account
// This is just for local API testing.

import (
	"context"
	"fmt"
	l "log"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/grupokindynos/go-binance"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		l.Println(err)
	}
}

func TestBinanceAPI(t *testing.T) {

	BINANCE_PUB_API := os.Getenv("BINANCE_PUB_API")
	BINANCE_PRIV_API := os.Getenv("BINANCE_PRIV_API")

	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, level.AllowAll())
	logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	hmacSigner := &binance.HmacSigner{
		Key: []byte(BINANCE_PRIV_API),
	}
	ctx, _ := context.WithCancel(context.Background())

	binanceService := binance.NewAPIService(
		"https://api.binance.com",
		BINANCE_PUB_API,
		hmacSigner,
		logger,
		ctx,
	)
	binanceApi := binance.NewBinance(binanceService)

	withdrawals, err := binanceApi.WithdrawHistory(binance.HistoryRequest{
		Asset:     "btc",
		Timestamp: time.Now(),
	})

	// withdrawal, err := binanceApi.Withdraw(binance.WithdrawRequest{
	// 	Asset:      strings.ToLower("DASH"),
	// 	Address:    "XtADvr2LwgNYLrkz3GPLddixZWJUFLEfTw",
	// 	Amount:     0.004,
	// 	Name:       "Adrestia-go Withdrawal",
	// 	RecvWindow: 5 * time.Second,
	// 	Timestamp:  time.Now(),
	// })
	// newOrder, err := binanceApi.NewOrder(binance.NewOrderRequest{
	// 	Symbol:    "DASHBTC",
	// 	Side:      binance.SideSell,
	// 	Type:      binance.TypeMarket,
	// 	Quantity:  0.2,
	// 	Timestamp: time.Now(),
	// })
	// res, err := binanceApi.QueryOrder(binance.QueryOrderRequest{
	// 	Symbol: "DASHBTC",
	// 	//OrderID:    126134994, BUY
	// 	//OrderID:    126136561,
	// 	OrderID:    126413548,
	// 	RecvWindow: 5 * time.Second,
	// 	//OrigClientOrderID: "b5JVkV0jQvAMqMuZwt0yp7",
	// 	Timestamp: time.Now(),
	// })

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, withdrawal := range withdrawals {
		fmt.Printf("%+v\n", withdrawal)
		if withdrawal.Id == "c1c9e3626b874350b2a0a07dfb3d3915" {
			fmt.Printf("%+v\n", withdrawal)
		}
	}
	//fmt.Println(newOrder)

	//fmt.Printf("%+v\n", res)
	//fmt.Printf("%+v\n", withdrawals)
}
