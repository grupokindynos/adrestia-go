package models

import (
	"fmt"
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