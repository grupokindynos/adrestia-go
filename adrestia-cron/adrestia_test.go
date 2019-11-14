package main

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/adrestia-cron/models"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/joho/godotenv"
	"log"
	"testing"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func TestOrdersAdrestia(t *testing.T) {
	var orderManager models.OrderManager
	orders := orderManager.GetOrderMap()
	fmt.Println(orders)
}

func TestOpenOrders(t *testing.T) {
	balOrders, err := services.GetBalancingOrders()
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("Balancing Orders: ", balOrders)
}