package processor

import (
	"errors"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/utils"
	"log"
	"time"
)

type ExchangesProcessor struct {
	Hestia services.HestiaService
	Plutus services.PlutusService
	Obol obol.ObolService
}

var(
	exchangesInfo []hestia.ExchangeInfo
	exchangeFactory *exchanges.ExchangeFactory
	newDeposits []hestia.SimpleTx
)

func (p *ExchangesProcessor) Start() {
	var err error
	deposits, err := p.Hestia.GetDeposits(false, 0)
	if err != nil {
		log.Println("ex_balancer- Start - unable to get deposits - " + err.Error())
	}
	if len(deposits) > 0 {
		log.Println("There are deposits still running")
		return
	}
	balancer, err := p.Hestia.GetBalancer()
	if err != nil {
		log.Println("ex_balancer - Start - unable to get balancer - " + err.Error())
	}
	emptyBalancer := hestia.Balancer{}
	if balancer != emptyBalancer {
		log.Println("There's a balancer still running")
		return
	}

	exchangesInfo, err = p.Hestia.GetExchanges()
	if err != nil {
		log.Println("ex_balancer - Start - Unable to get exchanges" + err.Error())
		return
	}
	if len(exchangesInfo) == 0 {
		log.Println("ex_balancer - Start - ExchangesInfo empty")
		return
	}
	exchangeFactory = exchanges.NewExchangeFactory(p.Obol, p.Hestia)
	err = p.Hestia.ChangeShiftProcessorStatus(false)
	if err != nil {
		log.Println("ex_balancer::Start::ChangeShiftProcessorStatus::" + err.Error())
		return
	}
	p.balanceExchanges()
}

func (p *ExchangesProcessor) balanceExchanges() {
	// Traer balances sin shifts
	balances, err := GetStockBalancesWithoutPendingShifts(p.Hestia, exchangesInfo, exchangeFactory)
	if err != nil {
		log.Println("ex_balancer::balanceExchanges::GetStockBalancesWithoutPendingShifts::" + err.Error())
		return
	}
	for _, exchangeInfo := range exchangesInfo {
		if bal, ok := balances[exchangeInfo.Name]; ok {
			if bal < exchangeInfo.StockMinimumAmount {
				err := p.createDeposit(exchangeInfo, exchangeInfo.StockExpectedAmount - bal)
				if err != nil {
					// This error is important, we should send a telegram message
					log.Println("ex_balancer - balanceExchanges - " + err.Error())
				}
			}
		}
	}
	
	isBalanceable, err := p.isDepositPossible()
	if err != nil {
		log.Println("ex_balancer - balanceExchanges - " + err.Error())
		return
	}
	
	if isBalanceable {
		for _, deposit := range newDeposits {
			_, err := p.Hestia.CreateDeposit(deposit)
			if err != nil {
				log.Println("ex_balancer - balanceExchanges - createDeposit - " + err.Error())
			}
		}
		err := p.Hestia.ChangeShiftProcessorStatus(true)
		if err != nil {
			log.Println("ex_balancer::balanceExchanges::ChangeShiftProcessorStatus::" + err.Error())
		}
	} else {
		balancer := hestia.Balancer{
			Id:            utils.RandomString(),
			Status:        hestia.BalancerStatusCreated,
			CreatedTime:   time.Now().Unix(),
			FulfilledTime: 0,
		}
		_, err := p.Hestia.CreateBalancer(balancer)
		if err != nil {
			log.Println("ex_balancer - balanceExchanges - createBalancer - " + err.Error())
		}
	}
}

func (p *ExchangesProcessor) createDeposit(exchangeInfo hestia.ExchangeInfo, amount float64) error {
	exchangeInstance, err := exchangeFactory.GetExchangeByName(exchangeInfo.Name, hestia.ShiftAccount)
	if err != nil {
		return err
	}

	addr, err := exchangeInstance.GetAddress(exchangeInfo.StockCurrency)
	if err != nil {
		return errors.New("createDeposit - " + err.Error())
	}

	if addr == "" {
		return errors.New("createDeposit - empty address returned")
	}

	deposit := hestia.SimpleTx{
		Id:         utils.RandomString(),
		TxId:       "",
		BalancerId: "",
		Exchange:   exchangeInfo.Name,
		Address:    addr,
		Currency:   exchangeInfo.StockCurrency,
		Amount:     amount,
		Status:     hestia.SimpleTxStatusCreated,
		CreatedTime:  time.Now().Unix(),
		FulfilledTime: 0,
	}
	newDeposits = append(newDeposits, deposit)
	return nil
}

func (p *ExchangesProcessor) isDepositPossible() (bool, error) {
	neededStock := make(map[string]float64)

	for _, deposit := range newDeposits {
		neededStock[deposit.Currency] += deposit.Amount
	}
	for currency, amount := range neededStock{
		balance, err := p.Plutus.GetWalletBalance(currency)
		if err != nil {
			// This message is also important
			return false, errors.New("isDepositPossible - unable to get balance for coin " + currency)
		}
		if balance.Confirmed < amount {
			return false, nil
		}
	}
	return true, nil
}
