package main

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/adrestia-go/utils"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	plutus2 "github.com/grupokindynos/common/plutus"
	"github.com/joho/godotenv"
)

var (
	testableCoins = [...]string{"BTC", "POLIS", "DASH"}

	orderPOLISDASH = hestia.AdrestiaOrder{
		ID:            "adrestia-test-polis-dash",
		DualExchange:  true,
		CreatedTime:   time.Now().Unix(),
		FulfilledTime: 0,
		Status:        hestia.AdrestiaStatusCreated,
		Amount:        20,
		BtcRate:       9684,
		FromCoin:      "POLIS",
		ToCoin:        "DASH",
		Message:       "",
		FirstOrder: hestia.ExchangeOrder{
			OrderId:          "",
			Symbol:           "POLISBTC",
			Side:             "sell",
			Amount:           0,
			ReceivedAmount:   0,
			CreatedTime:      0,
			FulfilledTime:    0,
			Exchange:         "southxchange",
			ReceivedCurrency: "BTC",
			SoldCurrency:     "POLIS",
		},
		FinalOrder: hestia.ExchangeOrder{
			OrderId:          "",
			Symbol:           "DASHBTC",
			Side:             "buy",
			Amount:           0,
			ReceivedAmount:   0,
			CreatedTime:      0,
			FulfilledTime:    0,
			Exchange:         "binance",
			ReceivedCurrency: "DASH",
			SoldCurrency:     "BTC",
		},
		HETxId:          "",
		EETxId:          "",
		EHTxId:          "",
		FirstExAddress:  "PRjoCA949ZpamrNpt9EU953zgCouC2mH3t",
		SecondExAddress: "157kMZrgThAmHrvinRLP4RKPC5AU4KdYKt",
		WithdrawAddress: "XuXMhik3Q5osBLqcWRj3JrNus3V5Ntnesd",
	}

	orderDASHPOLIS = hestia.AdrestiaOrder{
		ID:            "adrestia-test-dash-polis",
		DualExchange:  true,
		CreatedTime:   time.Now().Unix(),
		FulfilledTime: 0,
		Status:        hestia.AdrestiaStatusCreated,
		Amount:        0.5,
		BtcRate:       9684,
		FromCoin:      "DASH",
		ToCoin:        "POLIS",
		Message:       "",
		FirstOrder: hestia.ExchangeOrder{
			OrderId:          "",
			Symbol:           "DASHBTC",
			Side:             "sell",
			Amount:           0,
			ReceivedAmount:   0,
			CreatedTime:      0,
			FulfilledTime:    0,
			Exchange:         "binance",
			ReceivedCurrency: "BTC",
			SoldCurrency:     "DASH",
		},
		FinalOrder: hestia.ExchangeOrder{
			OrderId:          "",
			Symbol:           "POLISBTC",
			Side:             "buy",
			Amount:           0,
			ReceivedAmount:   0,
			CreatedTime:      0,
			FulfilledTime:    0,
			Exchange:         "southxchange",
			ReceivedCurrency: "POLIS",
			SoldCurrency:     "BTC",
		},
		HETxId:          "",
		EETxId:          "",
		EHTxId:          "",
		FirstExAddress:  "XuVmLDmUHZCjaSjm8KfXkGVhRG8fVC3Jis",
		SecondExAddress: "",
		WithdrawAddress: "PPrZjyLZqUazGaTGjrMmLMLg6FCRHVpypg",
	}

	orderBTCDASH = hestia.AdrestiaOrder{
		ID:            "adrestia-test-btc-dash",
		DualExchange:  false,
		CreatedTime:   time.Now().Unix(),
		FulfilledTime: 0,
		Status:        hestia.AdrestiaStatusCreated,
		Amount:        0.01,
		BtcRate:       9684,
		FromCoin:      "BTC",
		ToCoin:        "DASH",
		Message:       "",
		FirstOrder: hestia.ExchangeOrder{
			OrderId:          "",
			Symbol:           "DASHBTC",
			Side:             "buy",
			Amount:           0,
			ReceivedAmount:   0,
			CreatedTime:      0,
			FulfilledTime:    0,
			Exchange:         "binance",
			ReceivedCurrency: "DASH",
			SoldCurrency:     "BTC",
		},
		FinalOrder: hestia.ExchangeOrder{
			OrderId:          "",
			Symbol:           "",
			Side:             "",
			Amount:           0,
			ReceivedAmount:   0,
			CreatedTime:      0,
			FulfilledTime:    0,
			Exchange:         "",
			ReceivedCurrency: "",
			SoldCurrency:     "",
		},
		HETxId:          "",
		EETxId:          "",
		EHTxId:          "",
		FirstExAddress:  "157kMZrgThAmHrvinRLP4RKPC5AU4KdYKt",
		SecondExAddress: "",
		WithdrawAddress: "XuXMhik3Q5osBLqcWRj3JrNus3V5Ntnesd",
	}
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
}

func TestGetAddress(t *testing.T) {
	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	plutus := services.PlutusRequests{Obol: &oboli, PlutusURL: os.Getenv("PLUTUS_URL")}
	fmt.Println(plutus.GetAddress("POLIS"))
}

func TestSendToExchange(t *testing.T) {
	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	plutus := services.PlutusRequests{Obol: &oboli, PlutusURL: os.Getenv("PLUTUS_PRODUCTION_URL")}
	res, err := plutus.WithdrawToAddress(plutus2.SendAddressBodyReq{
		Address: "XjcPeT5AhGjoXdxuxwUwzaJE7SBbDKxyWG",
		Coin:    "DASH",
		Amount:  0.152670,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res)
}

func TestBalance(t *testing.T) {
	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	plutus := services.PlutusRequests{Obol: &oboli}
	bal, err := plutus.GetWalletBalance("BTC")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(bal)
}

func TestTestData(t *testing.T) {
	id, err := utils.CreateTestOrder(orderPOLISDASH)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(id)

	id, err = utils.CreateTestOrder(orderDASHPOLIS)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(id)
	id, err = utils.CreateTestOrder(orderBTCDASH)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(id)
}
