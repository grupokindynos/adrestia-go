package processor

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/grupokindynos/common/plutus"

	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models/adrestia"
	"github.com/grupokindynos/adrestia-go/services"
	cf "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
)

type Processor struct {
	Plutus          services.PlutusService
	Hestia          services.HestiaService
	Obol            obol.ObolService
	ExchangeFactory exchanges.IExchangeFactory
}

var (
	proc           Processor
	filledOrders   bool
	initialized    bool
	adrestiaOrders []hestia.AdrestiaOrder
)

func InitProcessor(params exchanges.Params) {
	proc.Plutus = params.Plutus
	proc.Hestia = params.Hestia
	proc.Obol = params.Obol
	proc.ExchangeFactory = params.ExchangeFactory
	initialized = true
}

func Start() {
	if !initialized {
		log.Println("error - Processor not initialized")
		return
	}
	const adrestiaStatus = true // TODO Replace with Hestia conf variable
	log.Println("Starting Adrestia Order Processor")

	if !adrestiaStatus {
		return
	}

	var wg sync.WaitGroup
	//wg.Add(5)
	wg.Add(1)
	//go handleCreatedOrders(&wg)
	//go handleExchange(&wg)
	go handleConversion(&wg)
	//go handleCompletedExchange(&wg)
	//go handleCompleted(&wg)
	wg.Wait()

	fmt.Println("Adrestia Order Processor Finished")
}

func handleCreatedOrders(wg *sync.WaitGroup) {
	defer wg.Done()
	orders := getOrders(hestia.AdrestiaStatusCreated)
	log.Println("CREATED_ORDERS", orders)
	for _, order := range orders {
		log.Println("entra")
		txId, err := proc.Plutus.WithdrawToAddress(plutus.SendAddressBodyReq{
			Address: order.FirstExAddress,
			Coin:    order.FromCoin,
			Amount:  order.Amount,
		})
		if err != nil {
			log.Println(fmt.Sprintf("error broadcasting order %s of coin %s: %v", order.ID, order.FromCoin, err))
			continue
		}
		order.HETxId = txId
		order.Status = hestia.AdrestiaStatusFirstExchange
		order.FirstOrder.CreatedTime = time.Now().Unix()
		_, err = proc.Hestia.UpdateAdrestiaOrder(order)
		if err != nil {
			log.Println(fmt.Sprintf("error updating order %s: %s", order.ID, err))
			continue
		}
	}
	fmt.Println("Finished CreatedOrders")
}

func handleExchange(wg *sync.WaitGroup) {
	defer wg.Done()
	firstExchangeOrders := getOrders(hestia.AdrestiaStatusFirstExchange)
	secondExchangeOrders := getOrders(hestia.AdrestiaStatusSecondExchange)
	for _, order := range firstExchangeOrders {
		ex, err := proc.ExchangeFactory.GetExchangeByName(order.FirstOrder.Exchange)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(order.FirstOrder.Exchange)
		status, err := ex.GetDepositStatus(order.HETxId, order.FromCoin)
		if err != nil {
			log.Println("117 " + err.Error())
			continue
		}
		log.Println("109")
		if status.Status == hestia.ExchangeStatusCompleted {
			order.FirstOrder.Amount = status.AvailableAmount
			_, orderId, err := ex.SellAtMarketPrice(order.FirstOrder)
			if err != nil {
				log.Println(err)
				continue
			}

			order.FirstOrder.OrderId = orderId
			order.Status = hestia.AdrestiaStatusFirstConversion
			updatedId, err := proc.Hestia.UpdateAdrestiaOrder(order)
			if err != nil {
				log.Println("HandleExchange: Failed to update order ", order.ID)
				continue
			}
			log.Println(fmt.Sprintf("HandleExchange: Successfully updated order %s to status %d", updatedId, hestia.AdrestiaStatusFirstConversion))
		}
	}

	for _, order := range secondExchangeOrders {
		ex, err := proc.ExchangeFactory.GetExchangeByName(order.FinalOrder.Exchange)
		if err != nil {
			log.Println(err)
			continue
		}
		status, err := ex.GetDepositStatus(order.EETxId, "BTC")
		if err != nil {
			log.Println("141 " + err.Error())
			continue
		}
		log.Println("140")
		if status.Status == hestia.ExchangeStatusCompleted {
			order.FinalOrder.Amount = status.AvailableAmount
			_, orderId, err := ex.SellAtMarketPrice(order.FinalOrder)
			if err != nil {
				log.Println(err)
				continue
			}

			order.FinalOrder.OrderId = orderId
			order.Status = hestia.AdrestiaStatusSecondConversion
			updatedId, err := proc.Hestia.UpdateAdrestiaOrder(order)
			if err != nil {
				log.Println("HandleExchange: Failed to update order ", order.ID)
				continue
			}
			log.Println(fmt.Sprintf("HandleExchange: Successfully updated order %s to status %d", updatedId, hestia.AdrestiaStatusFirstConversion))
		}
	}
	// 1. Verifies deposit in exchange and creates Selling Order always targets BTC
	fmt.Println("Finished handleExchange")
}

func handleConversion(wg *sync.WaitGroup) {
	defer wg.Done()

	ordersFirst := getOrders(hestia.AdrestiaStatusFirstConversion)
	ordersSecond := getOrders(hestia.AdrestiaStatusSecondConversion)

	orders := append(ordersFirst, ordersSecond...)
	var currExOrder *hestia.ExchangeOrder

	for _, order := range orders {
		if order.Status == hestia.AdrestiaStatusFirstConversion {
			currExOrder = &order.FirstOrder
		} else {
			currExOrder = &order.FinalOrder
		}
		exchange, err := proc.ExchangeFactory.GetExchangeByName(currExOrder.Exchange)
		if err != nil {
			fmt.Println(err)
			continue
		}

		status, err := exchange.GetOrderStatus(*currExOrder)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if status.Status == hestia.ExchangeStatusCompleted {
			currExOrder.FulfilledTime = time.Now().Unix()
			currExOrder.ReceivedAmount = status.AvailableAmount
			if err != nil {
				fmt.Println(err)
				continue
			}

			if order.DualExchange && order.Status == hestia.AdrestiaStatusFirstConversion {
				coin, err := cf.GetCoin(currExOrder.ReceivedCurrency)
				if err != nil {
					fmt.Println(err)
					continue
				}
				_, err = exchange.Withdraw(*coin, order.SecondExAddress, currExOrder.ReceivedAmount)
				if err != nil {
					fmt.Println(err)
					continue
				}
				order.FinalOrder.CreatedTime = time.Now().Unix()
				changeOrderStatus(order, hestia.AdrestiaStatusSecondExchange)
			} else {
				changeOrderStatus(order, hestia.AdrestiaStatusExchangeComplete)
			}
		} else if status.Status == hestia.ExchangeStatusError {
			changeOrderStatus(order, hestia.AdrestiaStatusError)
		}
	}

	// 1. Checks if order has been fulfilled.
	// 2. If target coin is BTC sends it to HW, else sends it to a second exchange
	fmt.Println("Finished handleConversion")
}

func handleCompletedExchange(wg *sync.WaitGroup) {
	defer wg.Done()
	orders := getOrders(hestia.AdrestiaStatusExchangeComplete)
	var exchangeOrder hestia.ExchangeOrder

	for _, order := range orders {
		if order.DualExchange {
			exchangeOrder = order.FinalOrder
		} else {
			exchangeOrder = order.FirstOrder
		}
		exchange, err := proc.ExchangeFactory.GetExchangeByName(exchangeOrder.Exchange)
		if err != nil {
			fmt.Println(err)
			continue
		}
		coin, err := cf.GetCoin(exchangeOrder.ReceivedCurrency)
		if err != nil {
			fmt.Println(err)
			continue
		}

		_, err = exchange.Withdraw(*coin, order.WithdrawAddress, exchangeOrder.ReceivedAmount)
		if err != nil {
			fmt.Println(err)
			continue
		}
		order.FulfilledTime = time.Now().Unix()
		changeOrderStatus(order, hestia.AdrestiaStatusCompleted)
	}

	fmt.Println("Finished handleCompletedExchange")
}

func handleCompleted(wg *sync.WaitGroup) {
	defer wg.Done()
	// Sends a telegram message and deletes order from CurrentOrders. Moves it to legacy table
	fmt.Println("Finished handleCompleted")
}

func changeOrderStatus(order hestia.AdrestiaOrder, status hestia.AdrestiaStatus) {
	fallbackStatus := order.Status
	order.Status = status
	resp, err := proc.Hestia.UpdateAdrestiaOrder(order)
	// TODO Move in map (if concurrency on maps allows for it)
	if err != nil {
		order.Status = fallbackStatus
		fmt.Println(err)
	} else {
		log.Println(fmt.Sprintf("order %s in %s has been updated to %d\t%s", order.FirstOrder.OrderId, order.FirstOrder.Exchange, order.Status, resp))
	}
}

func storeOrders(orders []hestia.AdrestiaOrder) {
	for _, order := range orders {
		res, err := proc.Hestia.CreateAdrestiaOrder(order)
		if err != nil {
			fmt.Println("error posting order to hestia: ", err)
		} else {
			fmt.Println(res)
		}
	}
}

func getOrders(status hestia.AdrestiaStatus) (filteredOrders []hestia.AdrestiaOrder) {
	if !filledOrders {
		res, err := fillOrders()
		if err != nil {
			log.Fatal(err)
			return
		}
		filledOrders = res
	}

	for _, order := range adrestiaOrders {
		if order.Status == status {
			filteredOrders = append(filteredOrders, order)
		}
	}
	return
}

func fillOrders() (bool, error) {
	var err error
	adrestiaOrders, err = proc.Hestia.GetAllOrders(adrestia.OrderParams{
		IncludeComplete: false,
	})

	if err != nil {
		log.Fatal("Could not retrieve adrestiaOrders form Hestia", err)
		return false, err
	}
	log.Println(fmt.Sprintf("Received a total of %d AdrestiaOrders", len(adrestiaOrders)))
	return true, nil
}
