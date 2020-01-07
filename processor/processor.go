package processor

import (
	"fmt"
	"log"
	"sync"

	"github.com/grupokindynos/adrestia-go/models/adrestia"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
)

type Processor struct {
	Hestia services.HestiaService
}

var adrestiaOrders []hestia.AdrestiaOrder

func (p *Processor) Start() {
	const adrestiaStatus = true // TODO Replace with Hestia conf variable
	log.Println("Starting Adrestia Order Processor")
	adrestiaOrders, err := p.Hestia.GetAllOrders(adrestia.OrderParams{
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
	wg.Add(6)
	go p.handleCreatedOrders(&wg)
	go p.handleFirstExchange(&wg)
	go p.handleSecondExchange(&wg)
	go p.handleFirstConversion(&wg)
	go p.handleCompletedExchange(&wg)
	go p.handleCompleted(&wg)
	wg.Wait()
	fmt.Println("Voucher Processor Finished")
}

func (p *Processor) handleCreatedOrders(wg *sync.WaitGroup) {
	defer wg.Done()
	orders := getOrders(hestia.AdrestiaStatusCreated)
	log.Println(orders)
	// 1.  Sends the amount to first exchange
	fmt.Println("Finished CreatedOrders")
}

func (p *Processor) handleFirstExchange(wg *sync.WaitGroup) {
	defer wg.Done()
	// 1. Verifies deposit in exchange and creates Selling Order always targets BTC
	fmt.Println("Finished SentAmount")
}

func (p *Processor) handleFirstConversion(wg *sync.WaitGroup) {
	defer wg.Done()
	// 1. Checks if order has been fulfilled.
	// 2. If target coin is BTC sends it to HW, else sends it to a second exchange
	fmt.Println("Finished SentAmount")
}

func (p *Processor) handleSecondExchange(wg *sync.WaitGroup) {
	defer wg.Done()
	// Verifies deposit in second exchange that targets the final coin. Arrives here if target is not BTC
	fmt.Println("Finished SentAmount")
}

func (p *Processor) ChangeOrderStatus(order hestia.AdrestiaOrder, status hestia.AdrestiaStatus) {
	fallbackStatus := order.Status
	order.Status = hestia.AdrestiaStatusStr[status]
	resp, err := p.Hestia.UpdateAdrestiaOrder(order)
	// TODO Move in map (if concurrency on maps allows for it)
	if err != nil {
		order.Status = fallbackStatus
		fmt.Println(err)
	} else {
		log.Println(fmt.Sprintf("order %s in %s has been updated to %s\t%s", order.OrderId, order.Exchange, order.Status, resp))
	}
}

func (p *Processor) StoreOrders(orders []hestia.AdrestiaOrder) {
	for _, order := range orders {
		res, err := p.Hestia.CreateAdrestiaOrder(order)
		if err != nil {
			fmt.Println("error posting order to hestia: ", err)
		} else {
			fmt.Println(res)
		}
	}
}
func (p *Processor) handleCompletedExchange(wg *sync.WaitGroup) {
	// Sends from final exchange to target coin HotWallet
}

func (p *Processor) handleCompleted(wg *sync.WaitGroup) {
	// Sends a telegram message and deletes order from CurrentOrders. Moves it to legacy table
}

func getOrders(status hestia.AdrestiaStatus) (filteredOrders []hestia.AdrestiaOrder) {
	for _, order := range adrestiaOrders {
		fmt.Println(order)
		if order.Status == hestia.AdrestiaStatusStr[status] {
			filteredOrders = append(filteredOrders, order)
		}
	}
	return
}
