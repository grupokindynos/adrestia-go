package processor

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/adrestia"
	"github.com/grupokindynos/adrestia-go/services"
	"log"
	"sync"
)

func Start() {
	const adrestiaStatus = true // TODO Replace with Hestia conf variable
	log.Println("Starting Adrestia Order Processor")

	adrestiaOrders, err := services.GetAllOrders(adrestia.OrderParams{
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
	go handleSentAmount(&wg)
	wg.Wait()
	fmt.Println("Voucher Processor Finished")

}

func handleCreatedOrders(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Finished CreatedOrders")
}

func handleSentAmount(wg *sync.WaitGroup) {
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
