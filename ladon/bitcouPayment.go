package ladon

import (
	"errors"
	"fmt"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/services"
	cf "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/explorer"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/telegram"
	"github.com/grupokindynos/common/utils"
	"github.com/joho/godotenv"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Unable to initialize .env " + err.Error())
	}
}

type BitcouPayment struct {
	Hestia services.HestiaService
	Obol   obol.ObolService
	ExFactory     *exchanges.ExchangeFactory
	ExInfo 		  []hestia.ExchangeInfo
	PaymentCoin   string
	PaymentAddress string
	BTCAddress	   string
	TgBot telegram.TelegramBot
}

var withdrawals []hestia.SimpleTx

func (bp *BitcouPayment) GenerateWithdrawals() {
	var paymentCoin string
	var paymentAddress string
	var totalBalanceUSD float64
	var totalBalanceUSDInfo float64
	balances := make(map[string]float64)

	rateBtcUsd, err := bp.Obol.GetCoin2FIATRate("BTC", "USD")
	if err != nil {
		log.Println("bitcouPayment::GenerateWithdrawals::GetCoin2FIATRate::" + err.Error())
		return
	}

	for _, exchange := range bp.ExInfo {
		if _, ok := BTCExchanges[exchange.Name]; ok { // exchanges where we should leave payment coin on BTC
			paymentCoin = "BTC"
		} else {
			paymentCoin = bp.PaymentCoin
		}

		exInstance, err := bp.ExFactory.GetExchangeByName(exchange.Name, hestia.VouchersAccount)
		if err != nil {
			log.Println("bitcouPayment::Start::GetExchangeByName::" + err.Error())
			continue
		}

		bal, err := exInstance.GetBalance(paymentCoin)
		if err != nil {
			log.Println("bitcouPayment::Start::GetBalance::" + exchange.Name + "::" + paymentCoin + "::" + err.Error())
			continue
		}

		balBTC := bal
		if paymentCoin == "BTC" {
			bal *= rateBtcUsd
		}
		// this value is used just to send the total balance on the telegram message.
		// shouldn't be used in any calculation
		totalBalanceUSDInfo += bal

		if bal >= minimumWithdrawalAmount {
			if paymentCoin == "BTC" {
				balances[exchange.Name] = balBTC
			} else {
				balances[exchange.Name] = bal
			}

			totalBalanceUSD += bal
		}
	}

	if totalBalanceUSD >= minimumPaymentAmount {
		for _, exchange := range bp.ExInfo {
			if _, ok := balances[exchange.Name]; !ok {continue}
			if _, ok := BTCExchanges[exchange.Name]; ok { // exchanges where we should leave payment coin on BTC
				paymentCoin = "BTC"
				paymentAddress = bp.BTCAddress

			} else {
				paymentCoin = bp.PaymentCoin
				paymentAddress = bp.PaymentAddress
			}

			exInstance, err := bp.ExFactory.GetExchangeByName(exchange.Name, hestia.VouchersAccount)
			if err != nil {
				log.Println("bitcouPayment::Start::GetExchangeByName::" + err.Error())
				continue
			}

			orderId, err := exInstance.Withdraw(paymentCoin, paymentAddress, balances[exchange.Name])
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
				Amount:         balances[exchange.Name],
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
		bp.TgBot.SendMessage(fmt.Sprintf("Total balance is less than the minimum payment amount.\nMinimum payment amount: %f USD\nTotal balance: %f USD", minimumPaymentAmount, totalBalanceUSDInfo), os.Getenv("BITCOU_CHAT_ID"))
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
	wg.Add(2)
	go bp.handleCreatedWithdrawals(&wg)
	go bp.handlePerformedWithdrawals(&wg)
	wg.Wait()
}

func (bp *BitcouPayment) handleCreatedWithdrawals(wg *sync.WaitGroup) {
	defer wg.Done()
	withdrawals := getWithdrawalsByStatus(hestia.SimpleTxStatusCreated)
	for _, withdrawal := range withdrawals {
		exchange, err := bp.ExFactory.GetExchangeByName(withdrawal.Exchange, hestia.VouchersAccount)
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

func (bp *BitcouPayment) handlePerformedWithdrawals(wg *sync.WaitGroup) {
	defer wg.Done()
	withdrawals := getWithdrawalsByStatus(hestia.SimpleTxStatusPerformed)
	for _, withdrawal := range withdrawals {
		amount, err := getBitcouReceivedAmount(withdrawal.Currency, withdrawal.Address, withdrawal.TxId)
		if err != nil {
			log.Println("bitcouPayment::getWithdrawalByStatus::getBitcouReceivedAmount::" + err.Error())
			continue
		}

		withdrawal.ReceivedAmount = amount
		withdrawal.Status = hestia.SimpleTxStatusCompleted
		withdrawal.FulfilledTime = time.Now().Unix()
		_, err = bp.Hestia.UpdateWithdrawal(withdrawal)
		if err != nil {
			log.Println("bitcouPayment::handleWithdrawals::UpdateWithdrawal::" + err.Error())
		}
		blockExplorer := ""
		coin, err := cf.GetCoin(withdrawal.Currency)
		if err != nil {
			blockExplorer = "Not Found"
		} else {
			blockExplorer = coin.Info.Blockbook
		}
		bp.TgBot.SendMessage(fmt.Sprintf("Sent %f %s at\n%s/tx/%s\nTo Bitcou Address %s", withdrawal.ReceivedAmount, withdrawal.Currency, blockExplorer, withdrawal.TxId, withdrawal.Address), os.Getenv("BITCOU_CHAT_ID"))
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

func getBitcouReceivedAmount(currency string, addr string, txId string) (float64, error) {
	token := false
	coin, err := cf.GetCoin(currency)
	if err != nil {
		return 0.0, errors.New("unable to get coin")
	}
	if coin.Info.Token && coin.Info.Tag != "ETH" {
		coin, _ = cf.GetCoin("ETH")
		token = true
	}
	blockbookWrapper := explorer.NewBlockBookWrapper(coin.Info.Blockbook)

	res, err := blockbookWrapper.GetTx(txId)
	if err != nil {
		return 0.0, errors.New("Error while getting tx " + err.Error())
	}

	if res.Confirmations > 0 {
		if token {
			for _, transfer := range res.TokenTransfers {
				if strings.ToLower(transfer.To) == strings.ToLower(addr) {
					value, _ := strconv.Atoi(transfer.Value)
					return float64(value) / math.Pow10(transfer.Decimals), nil
				}
			}
		} else {
			for _, txVout := range res.Vout {
				for _, address := range txVout.Addresses {
					if address == addr {
						value, err := strconv.ParseFloat(txVout.Value, 64)
						if err != nil {
							return 0.0, err
						}
						return value / math.Pow10(8), nil
					}
				}
			}
		}
	}

	return 0.0, errors.New("tx not found or still not confirmed")
}