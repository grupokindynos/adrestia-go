package main

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/common/hestia"
	"github.com/joho/godotenv"
	"log"
	"os"
	"testing"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
}

func TestBithumb(t *testing.T) {
	m := models.ExchangeParams{}
	m.Name = "Bithumb"
	m.Keys.PrivateKey = os.Getenv("BITHUMB_SECRET")
	m.Keys.PublicKey = os.Getenv("BITHUMB_API")

	b := exchanges.NewBithumb(m)

	assetBalance, err := b.GetBalance("USDT")
	if err != nil {
		return
	}
	fmt.Println(assetBalance)
}

func TestWithdrawBithumb(t *testing.T) {
	m := models.ExchangeParams{}
	m.Name = "Bithumb"
	m.Keys.PrivateKey = os.Getenv("BITHUMB_SECRET")
	m.Keys.PublicKey = os.Getenv("BITHUMB_API")

	b := exchanges.NewBithumb(m)

	assetBalance, err := b.Withdraw("USDT", "", 4)
	if err != nil {
		return
	}
	fmt.Println(assetBalance)
}

func TestMarketPrice(t *testing.T) {
	m := models.ExchangeParams{}
	m.Name = "Bithumb"
	m.Keys.PrivateKey = os.Getenv("BITHUMB_SECRET")
	m.Keys.PublicKey = os.Getenv("BITHUMB_API")

	b := exchanges.NewBithumb(m)

	assetBalance, err := b.SellAtMarketPrice(hestia.Trade{
		OrderId:        "",
		Amount:         0.900000,
		ReceivedAmount: 0,
		FromCoin:       "USDT",
		ToCoin:         "GTH",
		Symbol:         "GTH-USDT",
		Side:           "buy",
		Status:         0,
		Exchange:       "bithumb",
		CreatedTime:    0,
		FulfilledTime:  0,
	})
	if err != nil {
		return
	}
	fmt.Println(assetBalance)
}

func TestBithumbConfig(t *testing.T) {
	m := models.ExchangeParams{}
	m.Name = "Bithumb"
	m.Keys.PrivateKey = os.Getenv("BITHUMB_SECRET")
	m.Keys.PublicKey = os.Getenv("BITHUMB_API")

	b := exchanges.NewBithumb(m)

	config, err := b.GetPair("USDT", "GTH")
	if err != nil {
		return
	}
	fmt.Println(config)
}

func TestBithumbDeposits(t *testing.T) {
	m := models.ExchangeParams{}
	m.Name = "Bithumb"
	m.Keys.PrivateKey = os.Getenv("BITHUMB_SECRET")
	m.Keys.PublicKey = os.Getenv("BITHUMB_API")

	b := exchanges.NewBithumb(m)

	deposits, err := b.GetDepositStatus("0xb9cc9b046a901cff4a6943e26158c6de415a8b32", "0x5d18fb16b9a3dad6e1ff6cbc6dab3687a58269c2851325f1830b4c8c2d4d87ad", "USDT")
	if err != nil {
		return
	}
	fmt.Println(hestia.ExchangeOrderStatusStr[deposits.Status])
}