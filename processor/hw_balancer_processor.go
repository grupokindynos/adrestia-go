package processor

import (
	"errors"
	"github.com/grupokindynos/adrestia-go/balancer"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/utils"
	"log"
	"time"
)

type HwProcessor struct {
	Hestia services.HestiaService
	Plutus services.PlutusService
	Obol obol.ObolService
	Balancer balancer.Balancer
}

var(
	currentBalancer hestia.Balancer

	hwExchangesInfo []hestia.ExchangeInfo
	hwExFactory *exchanges.ExchangeFactory
)

func (hp *HwProcessor) Start() {
	var err error
	pendingWithdrawals, err := hp.Hestia.GetWithdrawals(false, 0, currentBalancer.Id)
	if err != nil {
		log.Println("hw_balancer - Start - Unable to get withdrawals " + err.Error())
		return
	}
	pendingBalancerOrders, err := hp.Hestia.GetBalanceOrders(false, 0)
	if err != nil {
		log.Println("hw_balancer - Start - balancerUnable to get balancer orders " + err.Error())
		return
	}
	currentBalancer, err = hp.Hestia.GetBalancer()
	if err != nil {
		log.Println("hw_balancer - Start - Unable to get balancer " + err.Error())
		return
	}
	hwExchangesInfo, err = hp.Hestia.GetExchanges()
	if err != nil {
		log.Println("hw_balancer - Start - Unable to get exchanges info " + err.Error())
		return
	}
	hwExFactory = exchanges.NewExchangeFactory(hp.Obol, hp.Hestia)
	processorParams := Params {
		Hestia: hp.Hestia,
		Plutus: hp.Plutus,
		Obol: hp.Obol,
	}
	withdrawalsProcessor := NewWithdrawalsProcessor(processorParams, pendingWithdrawals)
	balancerOrdersProcessor := NewBalancerOrderProcessor(processorParams, pendingBalancerOrders)

	switch currentBalancer.Status {
	case hestia.BalancerStatusCreated:
		hp.withdrawFromExchanges()
		currentBalancer.Status = hestia.BalancerStatusWithdrawal
		_, err := hp.Hestia.UpdateBalancer(currentBalancer)
		if err != nil {
			log.Println("hw_balancer - Start - StatusCreated - Unable to change status to balancer " + err.Error())
		}
		break
	case hestia.BalancerStatusWithdrawal:
		if len(pendingWithdrawals) > 0 {
			withdrawalsProcessor.Start()
		} else {
			err := hp.Balancer.Start(currentBalancer.Id)
			if err != nil {
				log.Println("hw_balancer - Start - StatusWithdrawal - balancer - " + err.Error())
				return
			}
			currentBalancer.Status = hestia.BalancerStatusTradeOrders
			_, err = hp.Hestia.UpdateBalancer(currentBalancer)
			if err != nil {
				log.Println("hw_balancer - Start - StatusWithdrawal - " + err.Error())
			}
		}
		break
	case hestia.BalancerStatusTradeOrders:
		if len(pendingBalancerOrders) > 0 {
			balancerOrdersProcessor.Start()
		} else {
			currentBalancer.Status = hestia.BalancerStatusCompleted
			currentBalancer.FulfilledTime = time.Now().Unix()
			_, err := hp.Hestia.UpdateBalancer(currentBalancer)
			if err != nil {
				log.Println("hw_balancer - Start - StatusTrade - updateBalancer" + err.Error())
			}
		}
	}
}

func (hp *HwProcessor) withdrawFromExchanges() {
	for _, exchangeInfo := range hwExchangesInfo {
		bal, err := getBalance(hwExFactory, exchangeInfo.Name, exchangeInfo.StockCurrency)
		if err != nil {
			log.Println("hw_balancer - withdrawFromExchanges - getBalance - " + err.Error())
			continue
		}
		if bal > exchangeInfo.StockMaximumAmount {
			err := hp.createWithdrawalOrder(exchangeInfo, bal - exchangeInfo.StockExpectedAmount)
			if err != nil {
				log.Println("hw_balancer - withdrawFromExchanges - createWithdrawalOrder " + err.Error())
			}
		}
	}
}

func (hp *HwProcessor) createWithdrawalOrder(exchangeInfo hestia.ExchangeInfo, amount float64) error {
	addr, err := hp.Plutus.GetAddress(exchangeInfo.StockCurrency)
	if err != nil {
		return errors.New("createWithdrawalOrder - GetAddress - " + err.Error())
	}
	withdrawal := hestia.SimpleTx{
		Id:             utils.RandomString(),
		TxId:           "",
		BalancerId:     currentBalancer.Id,
		Exchange:       exchangeInfo.Name,
		Address:        addr,
		Currency:       exchangeInfo.StockCurrency,
		Amount:         amount,
		ReceivedAmount: 0,
		Status:         hestia.SimpleTxStatusCreated,
		CreatedTime:    time.Now().Unix(),
		FulfilledTime:  0,
	}
	_, err = hp.Hestia.CreateWithdrawal(withdrawal)
	if err != nil {
		return errors.New("createWithdrawalOrder - CreateWithdrawal - " + err.Error())
	}
	return nil
}


