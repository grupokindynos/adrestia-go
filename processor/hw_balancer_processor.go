package processor

import (
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
	pendingWithdrawals, err := hp.Hestia.GetWithdrawals(false, 0)
	if err != nil {
		log.Println("Unable to get withdrawals " + err.Error())
		return
	}
	pendingBalancerOrders, err := hp.Hestia.GetBalanceOrders(false, 0)
	if err != nil {
		log.Println("Unable to get balancer orders" + err.Error())
		return
	}
	currentBalancer, err = hp.Hestia.GetBalancer()
	if err != nil {
		log.Println("Unable to get balancer " + err.Error())
		return
	}
	hwExchangesInfo, err = hp.Hestia.GetExchanges()
	if err != nil {
		log.Println("Unable to get exchanges info " + err.Error())
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
			log.Println("Unable to change status to balancer " + err.Error())
		}
	case hestia.BalancerStatusWithdrawal:
		if len(pendingWithdrawals) > 0 {
			withdrawalsProcessor.Start()
		} else {
			err := hp.Balancer.Start(currentBalancer.Id)
			if err != nil {
				log.Println("Balancer error: " + err.Error())
				return
			}
			currentBalancer.Status = hestia.BalancerStatusTradeOrders
			_, err = hp.Hestia.UpdateBalancer(currentBalancer)
			if err != nil {
				log.Println("Unable to change status to balancer " + err.Error())
			}
		}
	case hestia.BalancerStatusTradeOrders:
		if len(pendingBalancerOrders) > 0 {
			balancerOrdersProcessor.Start()
		} else {
			currentBalancer.Status = hestia.BalancerStatusCompleted
			currentBalancer.FulfilledTime = time.Now().Unix()
			_, err := hp.Hestia.UpdateBalancer(currentBalancer)
			if err != nil {
				log.Println("Unable to change status to balancer " + err.Error())
			}
		}
	default:
	}
}

func (hp *HwProcessor) withdrawFromExchanges() {
	for _, exchangeInfo := range hwExchangesInfo {
		bal, err := getBalance(hwExFactory, exchangeInfo.Name, exchangeInfo.StockCurrency)
		if err != nil {
			log.Println(err)
			continue
		}
		if bal > exchangeInfo.StockMaximumAmount {
			err := hp.createWithdrawalOrder(exchangeInfo, bal - exchangeInfo.StockExpectedAmount)
			if err != nil {
				log.Println("Error creating withdrawal order " + err.Error())
			}
		}
	}
}

func (hp *HwProcessor) createWithdrawalOrder(exchangeInfo hestia.ExchangeInfo, amount float64) error {
	addr, err := hp.Plutus.GetAddress(exchangeInfo.StockCurrency)
	if err != nil {
		return err
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
		return err
	}
	return nil
}


