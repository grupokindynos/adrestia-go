package processor

import (
	"errors"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/blockbook"
	cf "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/utils"
	"log"
	"strconv"
	"sync"
	"time"
)

type HwProcessor struct {
	Hestia services.HestiaService
	Plutus services.PlutusService
	Obol obol.ObolService
}

var(
	pendingWithdrawals[]hestia.SimpleTx
	pendingBalancerOrders[]hestia.BalancerOrder
	balancer hestia.Balancer

	hwExchangesInfo []hestia.ExchangeInfo
	hwExFactory *exchanges.ExchangeFactory
	blockExplorer blockbook.BlockBook
)

func (hp *HwProcessor) Start() {
	var err error
	pendingWithdrawals, err = hp.Hestia.GetWithdrawals(false, 0)
	if err != nil {
		log.Println("Unable to get withdrawals " + err.Error())
		return
	}
	pendingBalancerOrders, err = hp.Hestia.GetBalanceOrders(false, 0)
	if err != nil {
		log.Println("Unable to get balancer orders" + err.Error())
		return
	}
	balancer, err = hp.Hestia.GetBalancer()
	if err != nil {
		log.Println("Unable to get balancer " + err.Error())
		return
	}
	hwExchangesInfo, err = hp.Hestia.GetExchanges()
	if err != nil {
		log.Println("Unable to get exchanges info " + err.Error())
		return
	}
	hwExFactory = exchanges.NewExchangeFactory(hp.Obol)

	switch balancer.Status {
	case hestia.BalancerStatusCreated:
		hp.withdrawFromExchanges()
		balancer.Status = hestia.BalancerStatusWithdrawal
		hp.Hestia.UpdateBalancer(balancer)
	case hestia.BalancerStatusWithdrawal:
		if len(pendingWithdrawals) > 0 {
			hp.StartWithdrawalHandlers()
		} else{
			// Start balancer process
		}
	case hestia.BalancerStatusTradeOrders:
	default:
	}
}

func (hp *HwProcessor) StartWithdrawalHandlers() {
	var wg sync.WaitGroup
	wg.Add(3)
	go hp.handlerCreatedWithdrawal(&wg)
	go hp.handlerPerformedWithdrawal(&wg)
	go hp.handlerWithdrawalPlutusDeposit(&wg)
	wg.Wait()
}

func (hp *HwProcessor) handlerCreatedWithdrawal(wg *sync.WaitGroup) {
	defer wg.Done()
	withdrawals := getWithdrawalsByStatus(hestia.SimpleTxStatusCreated)
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
		_, err = hp.Hestia.UpdateWithdrawal(withdrawal)
		if err != nil {
			log.Println("Unable to update withdrawal on db " + err.Error())
		}
	}
}

func (hp *HwProcessor) handlerPerformedWithdrawal(wg *sync.WaitGroup) {
	defer wg.Done()
	withdrawals := getWithdrawalsByStatus(hestia.SimpleTxStatusPerformed)
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
			_, err := hp.Hestia.UpdateWithdrawal(withdrawal)
			if err != nil {
				log.Println("Error updating txId on withdrawal db " + err.Error())
			}
		}
	}
}

func (hp *HwProcessor) handlerWithdrawalPlutusDeposit(wg *sync.WaitGroup) {
	defer wg.Done()
	withdrawals := getWithdrawalsByStatus(hestia.SimpleTxStatusPlutusDeposit)
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
			withdrawal.Status = hestia.SimpleTxStatusCompleted
			_, err = hp.Hestia.UpdateWithdrawal(withdrawal)
			if err != nil {
				log.Println("Error while updating withdrawal on db " + err.Error())
			}
		}
	}
}

func (hp *HwProcessor) withdrawFromExchanges() {
	for _, exchangeInfo := range hwExchangesInfo {
		if exchangeInfo.StockAmount - exchangeInfo.StockExpectedAmount > 50 {
			err := hp.createWithdrawalOrder(exchangeInfo, exchangeInfo.StockAmount - exchangeInfo.StockExpectedAmount)
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
		BalancerId:     balancer.Id,
		Exchange:       exchangeInfo.Name,
		Address:        addr,
		Currency:       exchangeInfo.StockCurrency,
		Amount:         amount,
		ReceivedAmount: 0,
		Status:         hestia.SimpleTxStatusCreated,
		Timestamp:      time.Now().Unix(),
	}
	_, err = hp.Hestia.CreateWithdrawal(withdrawal)
	if err != nil {
		return err
	}
	return nil
}

func getPlutusReceivedAmount(tx blockbook.Tx, withdrawAddress string) (float64, error) {
	for _, txVout := range tx.Vout {
		for _, address := range txVout.Addresses {
			if address == withdrawAddress {
				value, err := strconv.ParseFloat(txVout.Value, 64)
				if err != nil {
					return 0.0, err
				}
				return value, nil
			}
		}
	}
	return 0.0, errors.New("Address not found")
}

func getWithdrawalsByStatus(status hestia.SimpleTxStatus) (filteredWithdrawals[]hestia.SimpleTx) {
	for _, withdrawal := range pendingWithdrawals{
		if withdrawal.Status == status {
			filteredWithdrawals = append(filteredWithdrawals, withdrawal)
		}
	}
	return
}

func getBalancerOrdersByStatus(status hestia.BalancerOrderStatus) (filteredOrders[]hestia.BalancerOrder) {
	for _, order := range  pendingBalancerOrders{
		if order.Status == status {
			filteredOrders = append(filteredOrders, order)
		}
	}
	return
}