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
	wg.Add(2)
	go p.handleCreatedOrders(&wg)
	go p.handleSentAmount(&wg)
	wg.Wait()
	fmt.Println("Voucher Processor Finished")

}

func (p *Processor) handleCreatedOrders(wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("Finished CreatedOrders")

}
func (p *Processor) handleSentAmount(wg *sync.WaitGroup) {
	defer wg.Done()

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
