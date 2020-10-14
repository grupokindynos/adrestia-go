package main

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/joho/godotenv"
	"log"
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
	m.Keys.PrivateKey = "3158feaf09b9d9181c596036db90563c394d9eff0224e4c4c5b2d0bd91aff44"
	m.Keys.PublicKey = "ee6819f1ec612258e743fc1aad71a7e0"

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
	m.Keys.PrivateKey = "3158feaf09b9d9181c596036db90563c394d9eff0224e4c4c5b2d0bd91aff44"
	m.Keys.PublicKey = "ee6819f1ec612258e743fc1aad71a7e0"

	b := exchanges.NewBithumb(m)

	assetBalance, err := b.Withdraw("USDT", "", 4)
	if err != nil {
		return
	}
	fmt.Println(assetBalance)
}