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
	wg.Add(5)
	go p.handleCreatedOrders(&wg)
	go p.handleExchange(&wg)
	go p.handleConversion(&wg)
	go p.handleCompletedExchange(&wg)
	go p.handleCompleted(&wg)
	wg.Wait()
	fmt.Println("Voucher Processor Finished")
}

func (p *Processor) handleCreatedOrders(wg *sync.WaitGroup) {
	defer wg.Done()
	orders := p.getOrders(hestia.AdrestiaStatusCreated)
	log.Println(orders)
	// 1.  Sends the amount to first exchange
	fmt.Println("Finished CreatedOrders")
}

func (p *Processor) handleExchange(wg *sync.WaitGroup) {
	defer wg.Done()
	// 1. Verifies deposit in exchange and creates Selling Order always targets BTC
	fmt.Println("Finished handleExchange")
}

func (p *Processor) handleConversion(wg *sync.WaitGroup) {
	defer wg.Done()
	// 1. Checks if order has been fulfilled.
	// 2. If target coin is BTC sends it to HW, else sends it to a second exchange
	fmt.Println("Finished handleConversion")
}

func (p *Processor) handleCompletedExchange(wg *sync.WaitGroup) {
	defer wg.Done()
	// Sends from final exchange to target coin HotWallet
	fmt.Println("Finished handleCompletedExchange")
}

func (p *Processor) handleCompleted(wg *sync.WaitGroup) {
	defer wg.Done()
	// Sends a telegram message and deletes order from CurrentOrders. Moves it to legacy table
	fmt.Println("Finished handleCompleted")
}

func (p *Processor) changeOrderStatus(order hestia.AdrestiaOrder, status hestia.AdrestiaStatus) {
	fallbackStatus := order.Status
	order.Status = status
	resp, err := p.Hestia.UpdateAdrestiaOrder(order)
	// TODO Move in map (if concurrency on maps allows for it)
	if err != nil {
		order.Status = fallbackStatus
		fmt.Println(err)
	} else {
		log.Println(fmt.Sprintf("order %s in %s has been updated to %s\t%s", order.FirstOrder.OrderId, order.FirstOrder.Exchange, order.Status, resp))
	}
}

func (p *Processor) storeOrders(orders []hestia.AdrestiaOrder) {
	for _, order := range orders {
		res, err := p.Hestia.CreateAdrestiaOrder(order)
		if err != nil {
			fmt.Println("error posting order to hestia: ", err)
		} else {
			fmt.Println(res)
		}
	}
}

func (p *Processor) getOrders(status hestia.AdrestiaStatus) (filteredOrders []hestia.AdrestiaOrder) {
	for _, order := range adrestiaOrders {
		fmt.Println(order)
		if order.Status == status {
			filteredOrders = append(filteredOrders, order)
		}
	}
	return
}
