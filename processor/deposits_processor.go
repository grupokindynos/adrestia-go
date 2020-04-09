package processor

import (
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/plutus"
	"log"
	"sync"
	"time"
)

type DepositProcessor struct {
	Hestia services.HestiaService
	Plutus services.PlutusService
	Obol obol.ObolService
}

var (
	pendingDeposits []hestia.SimpleTx
	depositstExFactory *exchanges.ExchangeFactory
)

func (p *DepositProcessor) Start() {
	var err error
	pendingDeposits, err = p.Hestia.GetDeposits(false, 0)
	if err != nil {
		log.Println("Unable to load deposits " + err.Error())
		return
	}

	depositstExFactory = exchanges.NewExchangeFactory(p.Obol, p.Hestia)

	var wg sync.WaitGroup
	wg.Add(2)
	go p.handleCreatedDeposit(&wg)
	go p.handlePerformedDeposit(&wg)
	wg.Wait()
}

func (p *DepositProcessor) handleCreatedDeposit(wg *sync.WaitGroup) {
	defer wg.Done()
	deposits := getDepositsByStatus(hestia.SimpleTxStatusCreated)
	for _, deposit := range deposits {
		txId, err := p.Plutus.WithdrawToAddress(plutus.SendAddressBodyReq{
			Address: deposit.Address,
			Coin:    deposit.Currency,
			Amount:  deposit.Amount,
		})
		if err != nil {
			log.Println("Unable to deposit to exchange " + err.Error())
			continue
		}
		deposit.TxId = txId
		deposit.Status = hestia.SimpleTxStatusPerformed
		_, err = p.Hestia.UpdateDeposit(deposit)
		if err != nil {
			log.Println("Unable to update deposit on db.", err.Error())
		}
	}
}

func (p *DepositProcessor) handlePerformedDeposit(wg *sync.WaitGroup) {
	defer wg.Done()
	deposits := getDepositsByStatus(hestia.SimpleTxStatusPerformed)
	for _, deposit := range deposits {
		exchange, err := depositstExFactory.GetExchangeByName(deposit.Exchange)
		if err != nil {
			log.Println("Unable to get exchange " + err.Error())
			continue
		}
		depositInfo, err := exchange.GetDepositStatus(deposit.Address, deposit.TxId, deposit.Currency)
		if err != nil {
			log.Println("Unable to get order status " + err.Error())
			continue
		}
		if depositInfo.Status == hestia.ExchangeOrderStatusCompleted {
			stock, err := exchange.GetBalance(deposit.Currency)
			if err != nil {
				log.Println("Unable to get new balance " + err.Error())
				continue
			}
			_, err = p.Hestia.UpdateExchangeBalance(deposit.Exchange, stock)
			if err != nil {
				log.Println("Unable to update new exchange balance on hestia " + err.Error())
				continue
			}
			deposit.ReceivedAmount = depositInfo.ReceivedAmount
			deposit.FulfilledTime = time.Now().Unix()
			deposit.Status = hestia.SimpleTxStatusCompleted
			_, err = p.Hestia.UpdateDeposit(deposit)
			if err != nil {
				log.Println("Unable to update deposit on db.", err.Error())
			}
		}
	}
}

func getDepositsByStatus(status hestia.SimpleTxStatus) (filteredDeposits []hestia.SimpleTx) {
	for _, deposit := range  pendingDeposits{
		if deposit.Status == status {
			filteredDeposits = append(filteredDeposits, deposit)
		}
	}
	return
}
