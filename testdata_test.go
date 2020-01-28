package main

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/adrestia-go/utils"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	plutus2 "github.com/grupokindynos/common/plutus"
	"github.com/joho/godotenv"
	"log"
	"os"
	"testing"
)

var (
	testableCoins = [...]string{"BTC", "POLIS", "DASH"}

	orderBTCPOLIS = hestia.AdrestiaOrder{
		ID:              "adrestia-test-btc-polis",
		DualExchange:    false,
		CreatedTime:     1579808853,
		FulfilledTime:   0,
		Status:          hestia.AdrestiaStatusCreated,
		Amount:          0.0001,
		BtcRate:         9684,
		FromCoin:        "BTC",
		ToCoin:          "POLIS",
		Message:         "",
		FirstOrder:      hestia.ExchangeOrder{
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
		FinalOrder:      hestia.ExchangeOrder{
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
		FirstExAddress:  "332gbBsyamxGvkxiFnkYdnvQ51bo2iEzHU",
		SecondExAddress: "",
		WithdrawAddress: "1FPPTjks4KRGssSi4c4EUZUcZKpMq3a73H",
	}

	orderDASHPOLIS = hestia.AdrestiaOrder{
		ID:              "adrestia-test-dash-polis",
		DualExchange:    true,
		CreatedTime:     1579809853,
		FulfilledTime:   0,
		Status:          hestia.AdrestiaStatusCreated,
		Amount:          0.01,
		BtcRate:         9684,
		FromCoin:        "DASH",
		ToCoin:          "POLIS",
		Message:         "",
		FirstOrder:      hestia.ExchangeOrder{
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
		FinalOrder:      hestia.ExchangeOrder{
			OrderId:          "",
			Symbol:           "POLISBTC",
			Side:             "",
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
		SecondExAddress: "332gbBsyamxGvkxiFnkYdnvQ51bo2iEzHU",
		WithdrawAddress: "PLauDHWLDMwoFn1TtWDHyw3jvt3V6qBGmw",
	}
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
}


func TestBtcAddress(t *testing.T) {
	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	plutus := services.PlutusRequests{Obol: &oboli}
	fmt.Println(plutus.GetAddress("polis"))
}

func TestSendToExchange(t *testing.T) {
	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	plutus := services.PlutusRequests{Obol: &oboli}
	fmt.Println(plutus.WithdrawToAddress(plutus2.SendAddressBodyReq{
		Address: "add address",
		Coin:    "DIVI",
		Amount:  2,
	}))
}

func TestBalance(t *testing.T) {
	oboli := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	plutus := services.PlutusRequests{Obol: &oboli}
	bal, err := plutus.GetWalletBalance("DGB")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(bal)
}

func TestTestData(t *testing.T) {
	id, err := utils.CreateTestOrder(orderBTCPOLIS)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(id)

	id, err = utils.CreateTestOrder(orderDASHPOLIS)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(id)
}
