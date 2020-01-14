package processor

import (
	"errors"
	"fmt"
	"github.com/grupokindynos/common/plutus"
	"log"
	"strings"
	"sync"

	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models/adrestia"
	"github.com/grupokindynos/adrestia-go/models/exchange_models"
	"github.com/grupokindynos/adrestia-go/services"
	cf "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
)

type Processor struct {
	Hestia services.HestiaService
	Obol   obol.ObolService
	Plutus services.PlutusRequests
}

var (
	adrestiaOrders  []hestia.AdrestiaOrder
	exchangeFactory *exchanges.ExchangeFactory
)

func (p *Processor) Start() {
	const adrestiaStatus = true // TODO Replace with Hestia conf variable
	log.Println("Starting Adrestia Order Processor")
	adrestiaOrders, err := p.Hestia.GetAllOrders(adrestia.OrderParams{
		IncludeComplete: false,
	})

	exchangeFactory = exchanges.NewExchangeFactory(exchange_models.Params{Obol: p.Obol})

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
	for _, order := range orders {
		txId, err := p.Plutus.WithdrawToAddress(plutus.SendAddressBodyReq{
			Address: order.FirstExAddress,
			Coin:    order.FromCoin,
			Amount:  order.Amount,
		})
		if err != nil {
			log.Println(fmt.Sprintf("error broadcasting order %s of coin %s", order.ID, order.FromCoin))
			continue
		}
		order.HETxId = txId
		order.Status = hestia.AdrestiaStatusFirstExchange
		_, err = p.Hestia.UpdateAdrestiaOrder(order)
		if err != nil {
			log.Println(fmt.Sprintf("error updating order %s", order.ID))
			continue
		}
	}
	fmt.Println("Finished CreatedOrders")
}

func (p *Processor) handleExchange(wg *sync.WaitGroup) {
	defer wg.Done()
	// 1. Verifies deposit in exchange and creates Selling Order always targets BTC
	fmt.Println("Finished handleExchange")
}

func (p *Processor) handleConversion(wg *sync.WaitGroup) {
	defer wg.Done()

	ordersFirst := p.getOrders(hestia.AdrestiaStatusFirstConversion)
	ordersSecond := p.getOrders(hestia.AdrestiaStatusSecondConversion)

	orders := append(ordersFirst, ordersSecond...)
	var currExOrder *hestia.ExchangeOrder

	for _, order := range orders {
		if order.Status == hestia.AdrestiaStatusFirstConversion {
			currExOrder = &order.FirstOrder
		} else {
			currExOrder = &order.FinalOrder
		}
		exchange, err := exchangeFactory.GetExchangeByName(currExOrder.Exchange)
		if err != nil {
			fmt.Println(err)
			continue
		}

		status, err := exchange.GetOrderStatus(order)
		if err != nil {
			fmt.Println(err)
			continue
		}

		obtainedCurrency, err := p.getObtainedCurrency(*currExOrder)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if status == hestia.ExchangeStatusCompleted {
			currExOrder.ListingAmount, err = exchange.GetListingAmount(*currExOrder)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if order.ToCoin == obtainedCurrency {
				p.changeOrderStatus(order, hestia.AdrestiaStatusCompleted)
			} else {
				coin, err := cf.GetCoin(obtainedCurrency)
				if err != nil {
					fmt.Println(err)
					continue
				}
				_, err = exchange.Withdraw(*coin, order.SecondExAddress, currExOrder.Amount)
				if err != nil {
					fmt.Println(err)
					continue
				}
				p.changeOrderStatus(order, hestia.AdrestiaStatusSecondExchange)
			}
		} else if status == hestia.ExchangeStatusError {
			p.changeOrderStatus(order, hestia.AdrestiaStatusError)
		}
	}

	// 1. Checks if order has been fulfilled.
	// 2. If target coin is BTC sends it to HW, else sends it to a second exchange
	fmt.Println("Finished handleConversion")
}

func (p *Processor) handleCompletedExchange(wg *sync.WaitGroup) {
	defer wg.Done()
	orders := p.getOrders(hestia.AdrestiaStatusCompleted)
	var exchangeOrder hestia.ExchangeOrder

	for _, order := range orders {
		if order.DualExchange {
			exchangeOrder = order.FinalOrder
		} else {
			exchangeOrder = order.FirstOrder
		}
		exchange, err := exchangeFactory.GetExchangeByName(exchangeOrder.Exchange)
		if err != nil {
			fmt.Println(err)
			continue
		}
		obtainedCurrency, err := p.getObtainedCurrency(exchangeOrder)
		if err != nil {
			fmt.Println(err)
			continue
		}
		coin, err := cf.GetCoin(obtainedCurrency)
		if err != nil {
			fmt.Println(err)
			continue
		}

		_, err = exchange.Withdraw(*coin, order.WithdrawAddress, exchangeOrder.ListingAmount)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

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

func (p *Processor) getObtainedCurrency(order hestia.ExchangeOrder) (string, error) {
	orderType := strings.ToLower(order.Side)

	if orderType == "buy" {
		return order.ReferenceCurrency, nil
	} else if orderType == "sell" {
		return order.ListingCurrency, nil
	} else {
		return "", errors.New("side not recognized " + orderType)
	}
}
