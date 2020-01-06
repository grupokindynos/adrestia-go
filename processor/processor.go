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

	fmt.Println("Finished SentAmount")
}
