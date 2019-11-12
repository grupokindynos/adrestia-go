package main

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/services"
	"testing"
)

func TestBalancing(t *testing.T) {
	var mdBalances []balance.Balance
	fmt.Println(mdBalances)
}

func TestOpenOrders(t *testing.T) {
	balOrders, err := services.GetBalancingOrders()
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("Balancing Orders: ", balOrders)
}