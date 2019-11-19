package models

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/api/exchanges"
	services2 "github.com/grupokindynos/adrestia-go/models/services"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
)

type OrderManager struct {

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