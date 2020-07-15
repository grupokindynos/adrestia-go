package ladon

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/utils"
	"log"
	"sync"
	"time"
)

type BitcouPayment struct {
	Hestia services.HestiaService
	Obol   obol.ObolService
	ExFactory     *exchanges.ExchangeFactory
	ExInfo 		  []hestia.ExchangeInfo
	PaymentCoin   string
	BTCExchanges  map[string]bool
	PaymentAddress string
	BTCAddress	   string
}

var withdrawals[] hestia.SimpleTx
var MINIMUM_PAYMENT_AMOUNT = 1000.0


func (bp *BitcouPayment) GenerateWithdrawals() {
	var paymentCoin string
	var paymentAddress string
	var totalBalanceUSD float64

	rateBtcUsd, err := bp.Obol.GetCoin2FIATRate("BTC", "USD")
	if err != nil {
		log.Println("bitcouPayment::GenerateWithdrawals::GetCoin2FIATRate::" + err.Error())
		return
	}

	for _, exchange := range bp.ExInfo {
		if _, ok := bp.BTCExchanges[exchange.Name]; ok { // exchanges where we should leave payment coin on BTC
			paymentCoin = "BTC"
		} else {
			paymentCoin = bp.PaymentCoin
		}

		exInstance, err := bp.ExFactory.GetExchangeByName(exchange.Name)
		if err != nil {
			log.Println("bitcouPayment::Start::GetExchangeByName::" + err.Error())
			continue
		}

		bal, err := exInstance.GetBalance(paymentCoin)
		if err != nil {
			log.Println("bitcouPayment::Start::GetBalance::" + err.Error())
			continue
		}

		if paymentCoin == "BTC" {
			bal *= rateBtcUsd
		}

		totalBalanceUSD += bal
	}

	if totalBalanceUSD >= MINIMUM_PAYMENT_AMOUNT {
		for _, exchange := range bp.ExInfo {
			if _, ok := bp.BTCExchanges[exchange.Name]; ok { // exchanges where we should leave payment coin on BTC
				paymentCoin = "BTC"
				paymentAddress = bp.BTCAddress

			} else {
				paymentCoin = bp.PaymentCoin
				paymentAddress = bp.PaymentAddress
			}

			exInstance, err := bp.ExFactory.GetExchangeByName(exchange.Name)
			if err != nil {
				log.Println("bitcouPayment::Start::GetExchangeByName::" + err.Error())
				continue
			}

			bal, err := exInstance.GetBalance(paymentCoin)
			if err != nil {
				log.Println("bitcouPayment::Start::GetBalance::" + err.Error())
				continue
			}

			orderId, err := exInstance.Withdraw(paymentCoin, paymentAddress, bal)
			if err != nil {
				log.Println("bitcouPayment::Start::Withdraw::" + err.Error())
				continue
			}

			withdrawal := hestia.SimpleTx {
				Id:             utils.RandomString(),
				TxId:           orderId,
				BalancerId:     "bitcouPayment",
				Exchange:       exchange.Name,
				Address:        paymentAddress,
				Currency:       paymentCoin,
				Amount:         bal,
				ReceivedAmount: 0,
				Status:         hestia.SimpleTxStatusCreated,
				CreatedTime:    time.Now().Unix(),
				FulfilledTime:  0,
			}

			_, err = bp.Hestia.CreateWithdrawal(withdrawal)
			if err != nil {
				log.Println("bitcouPayment::Start::CreateWithdrawal::" + err.Error())
			}
		}
	} else {
		log.Println(fmt.Sprintf("Total balance is less than the minimum payment amount.\nMinimum payment amount: %f USD\nTotal balance: %f USD", MINIMUM_PAYMENT_AMOUNT, totalBalanceUSD))
	}
}

func (bp *BitcouPayment) Start() {
	var err error
	withdrawals, err = bp.Hestia.GetWithdrawals(false, 0, "bitcouPayment")
	if err != nil {
		log.Println("bitcouPayment::Start::GetWithdrawals::" + err.Error())
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go bp.handleCreatedWithdrawals(&wg)
	wg.Wait()
}

func (bp *BitcouPayment) handleCreatedWithdrawals(wg *sync.WaitGroup) {
	defer wg.Done()
	withdrawals := getWithdrawalsByStatus(hestia.SimpleTxStatusCreated)

	for _, withdrawal := range withdrawals {
		exchange, err := bp.ExFactory.GetExchangeByName(withdrawal.Exchange)
		if err != nil {
			log.Println("bitcouPayment::handleWithdrawals::GetExchangeByName::" + err.Error())
			continue
		}
		txId, err := exchange.GetWithdrawalTxHash(withdrawal.TxId, withdrawal.Currency)
		if err != nil {
			log.Println("bitcouPayment::handleWithdrawals::GetWithdrawalsTxHash::" + err.Error())
			continue
		}
		if txId != "" {
			withdrawal.TxId = txId
			withdrawal.Status = hestia.SimpleTxStatusPerformed
			_, err := bp.Hestia.UpdateWithdrawal(withdrawal)
			if err != nil {
				log.Println("bitcouPayment::handleWithdrawals::UpdateWithdrawal::" + err.Error())
			}
		}
	}
}

func getWithdrawalsByStatus(status hestia.SimpleTxStatus) (filteredWithdrawals []hestia.SimpleTx) {
	for _, withdrawal := range withdrawals{
		if withdrawal.Status == status {
			filteredWithdrawals = append(filteredWithdrawals, withdrawal)
		}
	}
	return
}