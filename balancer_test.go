package main

import (
	"fmt"
	"testing"

	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/utils"
)

var (
	balanced = []balance.Balance{
		balance.Balance{
			Ticker:  "POLIS",
			RateBTC: 0.5,
			DiffBTC: 0.3,
		},
	}

	unbalanced = []balance.Balance{
		balance.Balance{
			Ticker:  "BTC",
			RateBTC: 1.0,
			DiffBTC: -0.1,
		},
		balance.Balance{
			Ticker:  "DASH",
			RateBTC: 0.03,
			DiffBTC: -0.15,
		},
	}
)

func TestBalanceHW(t *testing.T) {
	txs, _ := utils.BalanceHW(balanced, unbalanced)
	for _, tx := range txs {
		fmt.Printf("%+v\n", tx)
	}
}
