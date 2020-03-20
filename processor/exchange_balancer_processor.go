package processor

import (
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/utils"
	"log"
	"time"
)

type Processor struct {
	Hestia services.HestiaService
	Plutus services.PlutusService
	Obol obol.ObolService
}

var(
	exchangesInfo []hestia.ExchangeInfo
	exchangeFactory *exchanges.ExchangeFactory
	newDeposits []hestia.SimpleTx
)

func (p *Processor) Start() {
	var err error
	exchangesInfo, err = p.Hestia.GetExchanges()
	if err != nil {
		log.Println("Unable to get exchanges")
		return
	}
	if len(exchangesInfo) == 0 {
		log.Println("ExchangesInfo empty")
		return
	}
	exchangeFactory = exchanges.NewExchangeFactory(p.Obol)
	p.balanceExchanges()
}

func (p *Processor) balanceExchanges() {
	for _, exchangeInfo := range exchangesInfo {
		if exchangeInfo.StockAmount < exchangeInfo.StockMinimumAmount {
			err := p.createDeposit(exchangeInfo, exchangeInfo.StockExpectedAmount - exchangeInfo.StockAmount)
			if err != nil {
				// This error is important, we should send a telegram message
				log.Println(err)
			}
		}
	}
	
	isBalanceable, err := p.isDepositPossible()
	if err != nil {
		log.Println(err)
		return
	}
	
	if isBalanceable {
		for _, deposit := range newDeposits {
			_, err := p.Hestia.CreateDeposit(deposit)
			if err != nil {
				log.Println("unable to store deposit order " + err.Error())
			}
		}
	} else {

	}
}

func (p *Processor) createDeposit(exchangeInfo hestia.ExchangeInfo, amount float64) error {
	exchangeInstance, err := exchangeFactory.GetExchangeByName(exchangeInfo.Name)
	if err != nil {
		return err
	}

	addr, err := exchangeInstance.GetAddress(exchangeInfo.StockCurrency)
	if err != nil {
		return err
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
		Timestamp:  time.Now().Unix(),
	}
	newDeposits = append(newDeposits, deposit)
	return nil
}

func (p *Processor) isDepositPossible() (bool, error) {
	neededStock := make(map[string]float64)

	for _, deposit := range newDeposits {
		neededStock[deposit.Currency] += deposit.Amount
	}
	for currency, amount := range neededStock{
		balance, err := p.Plutus.GetWalletBalance(currency)
		if err != nil {
			// This message is also important
			log.Println("unable to get balance for coin " + currency)
			return false, err
		}
		if balance.Confirmed < amount {
			return false, nil
		}
	}
	return true, nil
}
