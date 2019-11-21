package models

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/api/exchanges"
	apiServices "github.com/grupokindynos/adrestia-go/api/services"
	services2 "github.com/grupokindynos/adrestia-go/models/services"
	"github.com/grupokindynos/adrestia-go/services"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"time"
)

type OrderManager struct {
	fiatThreshold float64 // USD // 2.0 for Testing, 10 USD for production
	orderTimeOut time.Duration // 2 * time.Hour
	exConfirmationThreshold int // 10
	walletConfirmationThreshold int // 3
	testingAmount float64 // 0.00001
}

func NewOrderManager(ft float64, ot time.Duration, ect int, wct int, ta float64) *OrderManager {
	o := new(OrderManager)
	o.fiatThreshold = ft
	o.orderTimeOut = ot
	o.exConfirmationThreshold = ect
	o.walletConfirmationThreshold = wct
	o.testingAmount = ta
	return o
}

func (o *OrderManager) GetOrderMap() map[string][]hestia.AdrestiaOrder {
	var adrestiaOrders = services2.AdrestiaOrderParams{
		IncludeComplete: true,
		AddedSince: 0,
	}
	orders, err := services.GetAllOrders(adrestiaOrders)
	if err!= nil {
		fmt.Println(err)
	}
	var mappedOrders = make(map[string][]hestia.AdrestiaOrder)
	for _, order := range orders {
		mappedOrders[order.Status] = append(mappedOrders[order.Status], order)
	}
	return mappedOrders
}

func (o *OrderManager) ExchangeCanFulfill(exchange exchanges.IExchange, coin string, amount float64) (bool, error) {
	var exAmount, err = exchange.GetBalances()
	if err != nil {
		return false, err
	}
	for _, asset := range exAmount {
		if asset.Ticker == coin {
			if asset.ConfirmedBalance > amount {
				return true, nil
			}
			if asset.ConfirmedBalance + asset.UnconfirmedBalance > amount {
				return false, nil
			}
		}
	}
	return false, nil
}

func (o *OrderManager)HandleBalances() {
	/*
		Fetches information about exchanges, their pending orders
	*/
}

func (o *OrderManager)HandleSentOrders(orders []hestia.AdrestiaOrder) {
	for _, order := range orders {
		tx, err :=services.GetWalletTx(order.FromCoin, order.TxId)
		if err != nil {
			fmt.Println(err)
		}
		if tx.Confirmations > o.exConfirmationThreshold {
			// TODO Create Order in Exchange and Update Status
			// TODO Handle rate variation
		}
	}
}

func (o *OrderManager)HandleCreatedOrders(orders []hestia.AdrestiaOrder) {
	ef := new(apiServices.ExchangeFactory)
	for _, order := range orders {
		coinInfo, _ := coinfactory.GetCoin(order.ToCoin)
		ex, err := ef.GetExchangeByCoin(*coinInfo)  // ex
		if err != nil {
			return
		}
		orderFulfilled := false
		// TODO ex.getOrderStatus
		if orderFulfilled {
			// TODO Withdraw
			conf, err := ex.Withdraw(*coinInfo, order.WithdrawAddress, 0.0)
			// conf, err := ex.Withdraw(*coinInfo, order.WithdrawAddress, order.Amount)
			if err != nil {
				fmt.Println(err)
				// TODO Bot report
				return
			}
			if conf {
				// TODO Update Status

			}

		}
	}
}

func (o *OrderManager)HandleWithdrawnOrders(orders []hestia.AdrestiaOrder) {
	for _, order := range orders {
		fmt.Println(order)
		// TODO Create exchange method for tracking order status
	}
}