package processor

import (
	"github.com/grupokindynos/adrestia-go/services"
	cf "github.com/grupokindynos/common/coin-factory"
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
			log.Println("unable to get exchange for withdrawal " + err.Error())
			continue
		}
		orderId, err := exchange.Withdraw(withdrawal.Currency, withdrawal.Address, withdrawal.Amount)
		if err != nil {
			log.Println("Error while trying to withdraw " + err.Error())
			continue
		}
		withdrawal.TxId = orderId
		withdrawal.Status = hestia.SimpleTxStatusPerformed
		_, err = wp.Hestia.UpdateWithdrawal(withdrawal)
		if err != nil {
			log.Println("Unable to update withdrawal on db " + err.Error())
		}
	}
}

func (wp *WithdrawalsProcessor) handlerPerformedWithdrawal(wg *sync.WaitGroup) {
	defer wg.Done()
	withdrawals := wp.getWithdrawalsByStatus(hestia.SimpleTxStatusPerformed)
	for _, withdrawal := range withdrawals {
		exchange, err := hwExFactory.GetExchangeByName(withdrawal.Exchange)
		if err != nil {
			log.Println("unable to get exchange for withdrawal txId " + err.Error())
			continue
		}
		txId, err := exchange.GetWithdrawalTxHash(withdrawal.TxId, withdrawal.Currency)
		if err != nil {
			log.Println("Error while getting txHash " + err.Error())
			continue
		}
		if txId != "" {
			withdrawal.TxId = txId
			withdrawal.Status = hestia.SimpleTxStatusPlutusDeposit
			_, err := wp.Hestia.UpdateWithdrawal(withdrawal)
			if err != nil {
				log.Println("Error updating txId on withdrawal db " + err.Error())
			}
		}
	}
}

func (wp *WithdrawalsProcessor) handlerWithdrawalPlutusDeposit(wg *sync.WaitGroup) {
	defer wg.Done()
	withdrawals := wp.getWithdrawalsByStatus(hestia.SimpleTxStatusPlutusDeposit)
	for _, withdrawal := range withdrawals {
		coin, err := cf.GetCoin(withdrawal.Currency)
		if err != nil {
			log.Println("Unable to get withdrawal coin " + err.Error())
			continue
		}
		blockExplorer.Url = "https://" + coin.BlockchainInfo.ExternalSource
		res, err := blockExplorer.GetTx(withdrawal.TxId)
		if err != nil {
			log.Println("Error while getting tx from blockbook " + err.Error())
			continue
		}
		if res.Confirmations > 0 {
			receivedAmount, err := getPlutusReceivedAmount(res, withdrawal.Address)
			if err != nil {
				log.Println("Error while getting blockbook received amount " + err.Error())
				continue
			}
			withdrawal.ReceivedAmount = receivedAmount
			withdrawal.FulfilledTime = time.Now().Unix()
			withdrawal.Status = hestia.SimpleTxStatusCompleted
			_, err = wp.Hestia.UpdateWithdrawal(withdrawal)
			if err != nil {
				log.Println("Error while updating withdrawal on db " + err.Error())
			}
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
