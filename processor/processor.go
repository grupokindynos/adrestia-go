package processor

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/adrestia"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"log"
	"sync"
)

var hestiaService = services.HestiaRequests{}
var adrestiaOrders[] hestia.AdrestiaOrder

func Start() {
	const adrestiaStatus = true // TODO Replace with Hestia conf variable

	log.Println("Starting Adrestia Order Processor")
	adrestiaOrders, err := hestiaService.GetAllOrders(adrestia.OrderParams{
		IncludeComplete: false,
	})

	if err != nil {
		log.Fatal("Could not retrieve adrestiaOrders form Hestia", err)
	}
	log.Println(fmt.Sprintf("Received a total of %d AdrestiaOrders", len(adrestiaOrders)))

	if !adrestiaStatus {
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go handleCreatedOrders(&wg)
	go handleFirstExchange(&wg)
	wg.Wait()
	fmt.Println("Voucher Processor Finished")
}

func handleCreatedOrders(wg *sync.WaitGroup) {
	defer wg.Done()
	orders := getOrders(hestia.AdrestiaStatusCreated)
	log.Println(orders)
	// 1.  Sends the amount to first exchange
	fmt.Println("Finished CreatedOrders")
}

func handleFirstExchange(wg *sync.WaitGroup) {
	defer wg.Done()
	// 1. Verifies deposit in exchange and creates Selling Order always targets BTC
	fmt.Println("Finished SentAmount")
}

func handleFirstConversion(wg *sync.WaitGroup) {
	defer wg.Done()
	// 1. Checks if order has been fulfilled.
	// 2. If target coin is BTC sends it to HW, else sends it to a second exchange
	fmt.Println("Finished SentAmount")
}

func handleSecondExchange(wg *sync.WaitGroup) {
	defer wg.Done()
	// Verifies deposit in second exchange that targets the final coin. Arrives here if target is not BTC
	fmt.Println("Finished SentAmount")
}

func handleConvertedCoins(wg *sync.WaitGroup) {
	// Sends from final exchange to target coin HotWallet
}

func handleCompletedOrders(wg *sync.WaitGroup) {
	// Sends a telegram message and deletes order from CurrentOrders. Moves it to legacy table
}

func getOrders(status hestia.AdrestiaStatus) (filteredOrders[] hestia.AdrestiaOrder){
	for _, order := range adrestiaOrders {
		fmt.Println(order)
		if order.Status == hestia.AdrestiaStatusStr[status] {
			filteredOrders = append(filteredOrders, order)
		}
	}
	return
}
