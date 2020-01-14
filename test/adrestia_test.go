package test

// import (
// 	"fmt"
// 	"github.com/grupokindynos/adrestia-go/models/order_manager"
// 	"github.com/grupokindynos/adrestia-go/services"
// 	"github.com/joho/godotenv"
// 	"github.com/stretchr/testify/assert"
// 	"log"
// 	"testing"
// )

// func init() {
// 	if err := godotenv.Load(); err != nil {
// 		log.Print("No .env file found")
// 	}
// }

// func TestOrdersAdrestia(t *testing.T) {
// 	var orderManager order_manager.OrderManager
// 	orders := orderManager.GetOrderMap()
// 	fmt.Println(orders)
// }

// func TestOpenOrders(t *testing.T) {
// 	balOrders, err := services.GetBalancingOrders()
// 	if err != nil {
// 		fmt.Print(err)
// 		return
// 	}
// 	fmt.Println("Balancing Orders: ", balOrders)
// }

// func TestPlutusTxInfo(t *testing.T) {
// 	tx, err := services.GetWalletTx("polis", "c28e88833169b0b383331beb9241c0db50c32911b2cabe32924ce3bdb150cc60")
// 	assert.Nil(t, err)
// 	assert.Equal(t, "05062aa74c43a95a268c3e50791bf4d1c644c061b53047dd9c3d5b0ea8e0240e", tx.Blockhash)
// 	assert.Equal(t, "c28e88833169b0b383331beb9241c0db50c32911b2cabe32924ce3bdb150cc60", tx.Txid)
// 	assert.Equal(t, 1163.39856061, tx.Vout[1].Value)
// }