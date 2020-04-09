package processor

import (
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/plutus"
	"log"
	"sync"
	"time"
)

type BalancerOrderProcessor struct {
	Hestia services.HestiaService
	Plutus services.PlutusService
	Obol   obol.ObolService
	balancerOrders []hestia.BalancerOrder
}

func NewBalancerOrderProcessor(params Params, orders []hestia.BalancerOrder) *BalancerOrderProcessor {
	return &BalancerOrderProcessor{
		Hestia:         params.Hestia,
		Plutus:         params.Plutus,
		Obol:			params.Obol,
		balancerOrders: orders,
	}
}

func (bp *BalancerOrderProcessor) Start() {
	var wg sync.WaitGroup
	wg.Add(5)
	go bp.handlerBalancerOrdersCreated(&wg)
	go bp.handlerBalancerOrdersExchangeDepositSent(&wg)
	go bp.handlerBalancerOrdersTrades(&wg)
	go bp.handlerBalancerOrdersWithdrawal(&wg)
	go bp.handlerBalancerOrdersPlutusDeposit(&wg)
	wg.Wait()
}

func (bp *BalancerOrderProcessor) handlerBalancerOrdersCreated(wg *sync.WaitGroup) {
	defer wg.Done()
	balancerOrders := bp.getBalancerOrdersByStatus(hestia.BalancerOrderStatusCreated)
	for _, order := range balancerOrders {
		txId, err := bp.Plutus.WithdrawToAddress(plutus.SendAddressBodyReq{
			Address: order.DepositAddress,
			Coin:    order.FromCoin,
			Amount:  order.OriginalAmount,
		})
		if err != nil {
			log.Println("Unable to withdraw to address " + err.Error())
			continue
		}
		order.Status = hestia.BalancerOrderStatusExchangeDepositSent
		order.DepositTxId = txId
		_, err = bp.Hestia.UpdateBalancerOrder(order)
		if err != nil {
			log.Println("Unable to update balancerOrder " + err.Error())
		}
	}
}

func (bp *BalancerOrderProcessor) handlerBalancerOrdersExchangeDepositSent(wg *sync.WaitGroup) {
	defer wg.Done()
	balancerOrders := bp.getBalancerOrdersByStatus(hestia.BalancerOrderStatusExchangeDepositSent)
	for _, order := range balancerOrders {
		exchange, err := hwExFactory.GetExchangeByName(order.Exchange)
		if err != nil {
			log.Println("Unable to get balancer by exchange")
			continue
		}
		depositInfo, err := exchange.GetDepositStatus(order.DepositAddress, order.DepositTxId, order.FromCoin)
		if err != nil {
			log.Println("Unable to get deposit info " + err.Error())
			continue
		}
		if depositInfo.Status == hestia.ExchangeOrderStatusCompleted {
			order.FirstTrade.Amount = depositInfo.ReceivedAmount
			order.FirstTrade.CreatedTime = time.Now().Unix()
			order.Status = hestia.BalancerOrderStatusFirstTrade
			orderId, err := exchange.SellAtMarketPrice(order.FirstTrade)
			if err != nil {
				log.Println("Error while placing trading order " + err.Error())
				continue
			}
			order.FirstTrade.OrderId = orderId
			_, err = bp.Hestia.UpdateBalancerOrder(order)
			if err != nil {
				log.Println("unable to update balancerOrder " + err.Error())
			}
		} else if depositInfo.Status == hestia.ExchangeOrderStatusError {
			log.Println("depositInfo status returned error status")
		}
	}
}

func (bp *BalancerOrderProcessor) handlerBalancerOrdersTrades(wg *sync.WaitGroup) {
	defer wg.Done()
	firstTrades := bp.getBalancerOrdersByStatus(hestia.BalancerOrderStatusFirstTrade)
	secondTrades := bp.getBalancerOrdersByStatus(hestia.BalancerOrderStatusSecondTrade)
	balancerOrders := append(firstTrades, secondTrades...)
	var tradeOrder *hestia.Trade
	for _, order := range balancerOrders {
		exchange, err := hwExFactory.GetExchangeByName(order.Exchange)
		if err != nil {
			log.Println("Unable to get exchange by name " + err.Error())
			continue
		}
		if order.Status == hestia.BalancerOrderStatusFirstTrade {
			tradeOrder = &order.FirstTrade
		} else {
			tradeOrder = &order.SecondTrade
		}
		orderInfo, err := exchange.GetOrderStatus(*tradeOrder)
		if err != nil {
			log.Println("Unable to get order status " + err.Error())
			continue
		}
		if orderInfo.Status == hestia.ExchangeOrderStatusCompleted {
			tradeOrder.ReceivedAmount = orderInfo.ReceivedAmount - 0.1
			tradeOrder.FulfilledTime = time.Now().Unix()
			if order.Status == hestia.BalancerOrderStatusFirstTrade && order.DualConversion {
				order.SecondTrade.Amount = orderInfo.ReceivedAmount
				order.SecondTrade.CreatedTime = time.Now().Unix()
				orderId, err := exchange.SellAtMarketPrice(order.SecondTrade)
				if err != nil {
					log.Println("Error placing second trading order " + err.Error())
					continue
				}
				order.SecondTrade.OrderId = orderId
				order.Status = hestia.BalancerOrderStatusSecondTrade
			} else {
				orderId, err := exchange.Withdraw(order.ToCoin, order.WithdrawAddress, tradeOrder.ReceivedAmount)
				if err != nil {
					log.Println("Error while trying to withdraw to hot wallet " + err.Error())
					continue
				}
				order.WithdrawTxId = orderId
				order.Status = hestia.BalancerOrderStatusWithdrawal
			}
			_, err := bp.Hestia.UpdateBalancerOrder(order)
			if err != nil {
				log.Println("Unable to update balancer order on db " + err.Error())
			}
		} else if orderInfo.Status ==  hestia.ExchangeOrderStatusError {
			log.Println("Order status returned error code")
		}
	}
}

func (bp *BalancerOrderProcessor) handlerBalancerOrdersWithdrawal(wg *sync.WaitGroup) {
	defer wg.Done()
	balancerOrders := bp.getBalancerOrdersByStatus(hestia.BalancerOrderStatusWithdrawal)
	for _, order := range balancerOrders {
		exchange, err := hwExFactory.GetExchangeByName(order.Exchange)
		if err != nil {
			log.Println("Unable to get exchange by name" + err.Error())
			continue
		}
		txId, err := exchange.GetWithdrawalTxHash(order.WithdrawTxId, order.ToCoin)
		if err != nil {
			log.Println("Error while getting txHash " + err.Error())
			continue
		}
		if txId != "" {
			order.WithdrawTxId = txId
			order.Status = hestia.BalancerOrderStatusPlutusDeposit
			_, err := bp.Hestia.UpdateBalancerOrder(order)
			if err != nil {
				log.Println("Unable to update withdrawTxId on db " + err.Error())
			}
		}
	}
}

func (bp *BalancerOrderProcessor) handlerBalancerOrdersPlutusDeposit(wg *sync.WaitGroup) {
	defer wg.Done()
	balancerOrders := bp.getBalancerOrdersByStatus(hestia.BalancerOrderStatusPlutusDeposit)
	for _, order := range balancerOrders {
		receivedAmount, err := getPlutusReceivedAmount(order.WithdrawAddress, order.WithdrawTxId)
		if err != nil {
			log.Println("Error while getting blockbook received amount " + err.Error())
			continue
		}
		order.ReceivedAmount = receivedAmount
		order.Status = hestia.BalancerOrderStatusCompleted
		order.FulfilledTime = time.Now().Unix()
		_, err = bp.Hestia.UpdateBalancerOrder(order)
		if err != nil {
			log.Println("Error while updating withdrawal on db " + err.Error())
		}
	}
}

func (bp *BalancerOrderProcessor) getBalancerOrdersByStatus(status hestia.BalancerOrderStatus) (filteredOrders[]hestia.BalancerOrder) {
	for _, order := range  bp.balancerOrders{
		if order.Status == status {
			filteredOrders = append(filteredOrders, order)
		}
	}
	return
}