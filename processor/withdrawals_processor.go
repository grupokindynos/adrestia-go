package processor

import (
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"log"
	"sync"
	"time"
)

type WithdrawalsProcessor struct {
	Hestia services.HestiaService
	Plutus services.PlutusService
	withdrawals []hestia.SimpleTx
}

func NewWithdrawalsProcessor(params Params, withdrawals []hestia.SimpleTx) *WithdrawalsProcessor {
	return &WithdrawalsProcessor {
		Hestia:      params.Hestia,
		Plutus:      params.Plutus,
		withdrawals: withdrawals,
	}
}

func (wp *WithdrawalsProcessor) Start() {
	var wg sync.WaitGroup
	wg.Add(3)
	go wp.handlerCreatedWithdrawal(&wg)
	go wp.handlerPerformedWithdrawal(&wg)
	go wp.handlerWithdrawalPlutusDeposit(&wg)
	wg.Wait()
}

func (wp *WithdrawalsProcessor) handlerCreatedWithdrawal(wg *sync.WaitGroup) {
	defer wg.Done()
	withdrawals := wp.getWithdrawalsByStatus(hestia.SimpleTxStatusCreated)
	for _, withdrawal := range withdrawals {
		exchange, err := hwExFactory.GetExchangeByName(withdrawal.Exchange)
		if err != nil {
			log.Println("withdrawals - handlerCreatedWithdrawal - GetExchangeByName - " + err.Error())
			continue
		}
		log.Println(withdrawal.Exchange)
		orderId, err := exchange.Withdraw(withdrawal.Currency, withdrawal.Address, withdrawal.Amount)
		if err != nil {
			log.Println("withdrawals - handlerCreatedWithdrawal - withdraw - " + err.Error())
			continue
		}
		withdrawal.TxId = orderId
		withdrawal.Status = hestia.SimpleTxStatusPerformed
		_, err = wp.Hestia.UpdateWithdrawal(withdrawal)
		if err != nil {
			log.Println("withdrawals - handlerCreatedWithdrawal - UpdateWithdrawal" + err.Error())
		}
	}
}

func (wp *WithdrawalsProcessor) handlerPerformedWithdrawal(wg *sync.WaitGroup) {
	defer wg.Done()
	withdrawals := wp.getWithdrawalsByStatus(hestia.SimpleTxStatusPerformed)
	for _, withdrawal := range withdrawals {
		exchange, err := hwExFactory.GetExchangeByName(withdrawal.Exchange)
		if err != nil {
			log.Println("withdrawals - handlerPerformedWithdrawal - GetExchangeByName - " + err.Error())
			continue
		}
		txId, err := exchange.GetWithdrawalTxHash(withdrawal.TxId, withdrawal.Currency)
		if err != nil {
			log.Println("withdrawals - handlerPerformedWithdrawals - GetWithdrawalsTxHash " + err.Error())
			continue
		}
		if txId != "" {
			withdrawal.TxId = txId
			withdrawal.Status = hestia.SimpleTxStatusPlutusDeposit
			_, err := wp.Hestia.UpdateWithdrawal(withdrawal)
			if err != nil {
				log.Println("withdrawals - handlerPerformedWithdrawals - UpdateWithdrawal - " + err.Error())
			}
		}
	}
}

func (wp *WithdrawalsProcessor) handlerWithdrawalPlutusDeposit(wg *sync.WaitGroup) {
	defer wg.Done()
	withdrawals := wp.getWithdrawalsByStatus(hestia.SimpleTxStatusPlutusDeposit)
	for _, withdrawal := range withdrawals {
		receivedAmount, err := getPlutusReceivedAmount(withdrawal.Address, withdrawal.TxId)
		if err != nil {
			log.Println("withdrawals - handlerWithdrawalPlutusDeposit - getReceivedAmount - " + err.Error())
			continue
		}
		withdrawal.ReceivedAmount = receivedAmount
		withdrawal.FulfilledTime = time.Now().Unix()
		withdrawal.Status = hestia.SimpleTxStatusCompleted
		_, err = wp.Hestia.UpdateWithdrawal(withdrawal)
		if err != nil {
			log.Println("withdrawals - handlerWithdrawalPlutusDeposit - UpdateWithdrawal - " + err.Error())
		}
	}
}

func (wp *WithdrawalsProcessor) getWithdrawalsByStatus(status hestia.SimpleTxStatus) (filteredWithdrawals[]hestia.SimpleTx) {
	for _, withdrawal := range wp.withdrawals{
		if withdrawal.Status == status {
			filteredWithdrawals = append(filteredWithdrawals, withdrawal)
		}
	}
	return
}
