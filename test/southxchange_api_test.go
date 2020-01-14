package test

import (
	"fmt"
	south "github.com/oedipusK/go-southxchange"
	"testing"
)

func TestAPI(t *testing.T) {
	key := "gKCeKnJqjoUdIXKfwnmPEXgFhQscIT"
	secret := "YQIFLfwnPUovabquVyeFwavgUIKcVabMHOoDTiLkIoWqgNdlTs"
	//orderId := "WR8VJR54"
	southClient := *south.New(key, secret, "user-agent")

	// order, err := southClient.GetOrder(orderId)
	// if err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	txs, err := southClient.GetTransactions(0, 500, "", true)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, tx := range txs {
		fmt.Printf("%+v\n", tx)
	}
}
