package processor

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models/adrestia"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/adrestia-go/telegram"
	blockbook "github.com/grupokindynos/common/blockbook"
	cf "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/plutus"
)

type Processor struct {
	Plutus          services.PlutusService
	Hestia          services.HestiaService
	Obol            obol.ObolService
	ExchangeFactory exchanges.IExchangeFactory
}

var (
	proc           Processor
	initialized    bool
	adrestiaOrders []hestia.AdrestiaOrder
	blockExplorer  blockbook.BlockBook
	telegramBot    = telegram.NewTelegramBot()
)

func InitProcessor(params exchanges.Params) {
	proc.Plutus = params.Plutus
	proc.Hestia = params.Hestia
	proc.Obol = params.Obol
	proc.ExchangeFactory = params.ExchangeFactory
	initialized = true
}

func Start() {
	status, err := proc.Hestia.GetAdrestiaStatus()
	if err != nil {
		log.Println("Couldn't get adrestia status" + err.Error())
		return
	}
	if !status.Service {
		log.Println("Processor not available at the moment")
		return
	}
	if !initialized {
		log.Println("error - Processor not initialized")
		return
	}

	log.Println("Starting Adrestia Order Processor")

	err = fillOrders()
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	wg.Add(6)
	//wg.Add(1)
	go handleCreatedOrders(&wg)
	go handleExchange(&wg)
	go handleConversion(&wg)
	go handleWithdrawal(&wg)
	go handleCompletedExchange(&wg)
	go handlePlutusDeposit(&wg)
	wg.Wait()

	fmt.Println("Adrestia Order Processor Finished")
}

func handleCreatedOrders(wg *sync.WaitGroup) {
	defer wg.Done()
	orders := getOrders(hestia.AdrestiaStatusCreated)
	for _, order := range orders {
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
			log.Println("handleExchange - GetExchangeByName() - " + err.Error())
			continue
		}
		status, err := ex.GetDepositStatus(order.HETxId, order.FromCoin)
		if err != nil {
			log.Println("handleExchange - GetDepositStatus() - " + err.Error())
			continue
		}
		if status.Status == hestia.ExchangeStatusCompleted {
			order.FirstOrder.Amount = status.AvailableAmount
			orderId, err := ex.SellAtMarketPrice(order.FirstOrder)
			if err != nil {
				log.Println("handleExchange - SellAtMarketPrice() - " + err.Error())
				continue
			}

			order.FirstOrder.OrderId = orderId
			order.Status = hestia.AdrestiaStatusFirstConversion
			updatedId, err := proc.Hestia.UpdateAdrestiaOrder(order)
			if err != nil {
				log.Println("handleExchange - UpdateAdrestiaOrder - ", order.ID, " - "+err.Error())
				continue
			}
			log.Println(fmt.Sprintf("HandleExchange: Successfully updated order %s to status %d", updatedId, hestia.AdrestiaStatusFirstConversion))
		}
	}

	for _, order := range secondExchangeOrders {
		var status hestia.OrderStatus
		var ex exchanges.IExchange
		var err error
		// Check if all the trading process is going to be done on the same exchange
		if order.FirstOrder.Exchange == order.FinalOrder.Exchange {
			status.Status = hestia.ExchangeStatusCompleted
			status.AvailableAmount = order.FirstOrder.ReceivedAmount
		} else {
			ex, err = proc.ExchangeFactory.GetExchangeByName(order.FinalOrder.Exchange)
			if err != nil {
				log.Println("handleExchange - GetExchangeByName() - " + err.Error())
				continue
			}
			status, err = ex.GetDepositStatus(order.EETxId, "BTC")
			if err != nil {
				log.Println("handleExchange - 2nd GetDepositStatus() - " + err.Error())
				continue
			}
		}
		if status.Status == hestia.ExchangeStatusCompleted {
			order.FinalOrder.Amount = status.AvailableAmount
			orderId, err := ex.SellAtMarketPrice(order.FinalOrder)
			if err != nil {
				log.Println("handleExchange - 2nd SellAtMarketPrice() - " + err.Error())
				continue
			}

			order.FinalOrder.OrderId = orderId
			order.Status = hestia.AdrestiaStatusSecondConversion
			updatedId, err := proc.Hestia.UpdateAdrestiaOrder(order)
			if err != nil {
				log.Println("HandleExchange: Failed to update order ", order.ID)
				continue
			}
			log.Println(fmt.Sprintf("HandleExchange: Successfully updated order %s to status %d", updatedId, hestia.AdrestiaStatusSecondConversion))
		}
	}
	// 1. Verifies deposit in exchange and creates Selling Order always targets BTC
	fmt.Println("Finished handleExchange")
}

func handleWithdrawal(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Start handleWithdrawal")

	ordersFirstWithdrawal := getOrders(hestia.AdrestiaStatusFirstWithdrawal) // waiting for withdrawal
	ordersSecondWithdrawal := getOrders(hestia.AdrestiaStatusSecondWithdrawal)

	orders := append(ordersFirstWithdrawal, ordersSecondWithdrawal...)

	var currExOrder *hestia.ExchangeOrder
	var withdrawalId string
	var withdrawalAddress string
	var toExchange bool
	var nextState hestia.AdrestiaStatus

	for _, order := range orders {
		if order.Status == hestia.AdrestiaStatusFirstWithdrawal {
			currExOrder = &order.FirstOrder
			if order.DualExchange {
				toExchange = true
				withdrawalId = order.EETxId
				withdrawalAddress = order.SecondExAddress
				nextState = hestia.AdrestiaStatusSecondExchange
			} else {
				withdrawalId = order.EHTxId
				withdrawalAddress = order.WithdrawAddress
				nextState = hestia.AdrestiaStatusPlutusDeposit
			}
		} else {
			currExOrder = &order.FinalOrder
			withdrawalId = order.EHTxId
			withdrawalAddress = order.WithdrawAddress
			nextState = hestia.AdrestiaStatusPlutusDeposit
		}

		coin, err := cf.GetCoin(currExOrder.ReceivedCurrency)
		if err != nil {
			fmt.Println("handleWithdrawal - GetCoin() - " + err.Error())
			continue
		}

		exchange, err := proc.ExchangeFactory.GetExchangeByName(currExOrder.Exchange)
		if err != nil {
			fmt.Println("handleWithdrawal - GetExchangeByName() - " + err.Error())
			continue
		}

		txHash, err := exchange.GetWithdrawalTxHash(withdrawalId, coin.Info.Tag, withdrawalAddress, currExOrder.ReceivedAmount)
		if err != nil {
			log.Println("handleWithdrawal - GetWithdrawalTxHash() - " + err.Error())
			continue
		}
		if txHash != "" {
			if toExchange {
				order.EETxId = txHash
			} else {
				order.EHTxId = txHash
			}
			changeOrderStatus(order, nextState)
		}
	}
	log.Println("handleWithdrawal completed")
}

func handleConversion(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Starting handleConversion")

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
			fmt.Println("handleConversion - GetExchangeByName() - " + err.Error())
			continue
		}

		status, err := exchange.GetOrderStatus(*currExOrder)
		if err != nil {
			fmt.Println("handleConversion - GetOrderStatus() - ", err.Error())
			continue
		}

		if status.Status == hestia.ExchangeStatusCompleted {
			currExOrder.FulfilledTime = time.Now().Unix()
			currExOrder.ReceivedAmount = status.AvailableAmount

			if order.DualExchange && order.Status == hestia.AdrestiaStatusFirstConversion {
				// Check if the second conversion needs to be done on another exchange.
				if order.FirstOrder.Exchange != order.FinalOrder.Exchange {
					coin, err := cf.GetCoin(currExOrder.ReceivedCurrency)
					if err != nil {
						fmt.Println("handleConversion - GetCoin() - ", err.Error())
						continue
					}
					txid, err := exchange.Withdraw(*coin, order.SecondExAddress, currExOrder.ReceivedAmount)
					if err != nil {
						fmt.Println("handleConversion - Withdraw() - " + currExOrder.Exchange + " " + err.Error())
						continue
					}
					order.EETxId = txid
					order.FinalOrder.CreatedTime = time.Now().Unix()
					changeOrderStatus(order, hestia.AdrestiaStatusFirstWithdrawal)
				} else {
					order.FinalOrder.CreatedTime = time.Now().Unix()
					changeOrderStatus(order, hestia.AdrestiaStatusSecondExchange)
				}
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

	log.Println("Starting handleCompletedExchange")
	orders := getOrders(hestia.AdrestiaStatusExchangeComplete)
	var exchangeOrder hestia.ExchangeOrder
	var nextState hestia.AdrestiaStatus

	for _, order := range orders {
		if order.DualExchange {
			exchangeOrder = order.FinalOrder
			nextState = hestia.AdrestiaStatusSecondWithdrawal
		} else {
			exchangeOrder = order.FirstOrder
			nextState = hestia.AdrestiaStatusFirstWithdrawal
		}
		exchange, err := proc.ExchangeFactory.GetExchangeByName(exchangeOrder.Exchange)
		if err != nil {
			fmt.Println("handleCompletedExchange - GetExchangeByName() - ", err.Error())
			continue
		}
		coin, err := cf.GetCoin(exchangeOrder.ReceivedCurrency)
		if err != nil {
			fmt.Println("handleCompletedExchange - GetCoin() - ", err.Error())
			continue
		}

		txId, err := exchange.Withdraw(*coin, order.WithdrawAddress, exchangeOrder.ReceivedAmount)
		if err != nil {
			fmt.Println("handleCompletedExchange - Withdraw() - ", err.Error())
			continue
		}

		order.EHTxId = txId
		changeOrderStatus(order, nextState)
	}

	fmt.Println("Finished handleCompletedExchange")
}

func handlePlutusDeposit(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Start handlePlutusDeposit")

	orders := getOrders(hestia.AdrestiaStatusPlutusDeposit)
	for _, order := range orders {
		coin, err := cf.GetCoin(order.ToCoin)
		if err != nil {
			log.Println("handlePlutusDeposit - GetCoin() - ", err.Error())
			continue
		}

		blockExplorer.Url = "https://" + coin.BlockchainInfo.ExternalSource
		res, err := blockExplorer.GetTx(order.EHTxId)
		if err != nil {
			log.Println("handlePlutusDeposit - GetTx() - ", err.Error())
			continue
		}
		if res.Confirmations > 0 {
			receivedAmount, err := getReceivedAmount(res, order.WithdrawAddress)
			if err != nil {
				log.Println("handlePlutusDeposit - getReceivedAmount() - ", err.Error())
				continue
			}
			order.ReceivedAmount = receivedAmount
			order.FulfilledTime = time.Now().Unix()
			rate, err := proc.Obol.GetCoin2CoinRates(order.FromCoin, order.ToCoin)
			if err != nil {
				log.Println("handlePlutusDeposit - GetCoin2CoinRates() - ", err.Error())
				rate = 1.0
			}
			telegramBot.SendMessage(fmt.Sprintf("Change from %s to %s has been completed.\nSent %.8f %s and received %.8f %s (~%.8f %s).\nOrderId: %s", order.FromCoin, order.ToCoin, order.Amount, order.FromCoin, order.ReceivedAmount*1e-8, order.ToCoin, order.ReceivedAmount*1e-8/rate, order.FromCoin, order.ID))
			changeOrderStatus(order, hestia.AdrestiaStatusCompleted)
		}
	}
	log.Println("Finished handlePlutusDeposit")
}

func getReceivedAmount(tx blockbook.Tx, withdrawAddress string) (float64, error) {
	for _, txVout := range tx.Vout {
		for _, address := range txVout.Addresses {
			if address == withdrawAddress {
				value, err := strconv.ParseFloat(txVout.Value, 64)
				if err != nil {
					return 0.0, err
				}
				return value, nil
			}
		}
	}
	return 0.0, errors.New("Address not found")
}

func changeOrderStatus(order hestia.AdrestiaOrder, status hestia.AdrestiaStatus) {
	fallbackStatus := order.Status
	order.Status = status
	resp, err := proc.Hestia.UpdateAdrestiaOrder(order)
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
	for _, order := range adrestiaOrders {
		if order.Status == status {
			filteredOrders = append(filteredOrders, order)
		}
	}
	return
}

func fillOrders() error {
	var err error
	adrestiaOrders, err = proc.Hestia.GetAllOrders(adrestia.OrderParams{
		IncludeComplete: false,
	})

	if err != nil {
		log.Fatal("Could not retrieve adrestiaOrders form Hestia", err)
		return err
	}
	log.Println(fmt.Sprintf("Received a total of %d AdrestiaOrders", len(adrestiaOrders)))
	return nil
}
