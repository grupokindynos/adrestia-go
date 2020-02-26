package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	south "github.com/oedipusK/go-southxchange"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
}

func TestSouthAPI(t *testing.T) {
	key := os.Getenv("SOUTH_API_KEY")
	secret := os.Getenv("SOUTH_API_SECRET")
	//orderId := "WR8VJR54"
	southClient := *south.New(key, secret, "user-agent")

	// order, err := southClient.GetOrder(orderId)
	// if err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	txs, err := southClient.GetTransactions("", 0, 500, "", true)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, tx := range txs {
		fmt.Printf("%+v\n", tx)
	}
	//3ANOQ16U

	//price, err := southClient.GetMarketPrice("POLIS", "BTC")

	// addr, err := southClient.GetDepositAddress("BTC")
	// if err != nil {
	// 	fmt.Println("Error " + err.Error())
	// 	return
	// }
	// fmt.Println(addr)
	//addr, err = southClient.GetDepositAddress("BTC")
	// if err != nil {
	// 	fmt.Println("Error " + err.Error())
	// 	return
	// }
	// for _, tx := range txs {
	// 	fmt.Printf("%+v\n", tx)
	// }
	//fmt.Println(addr)
	//fmt.Printf("%+v\n", price)
}
