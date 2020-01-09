package handlers

import (
	"fmt"
	"time"

	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models/adrestia"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/services"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/plutus"
	"github.com/lithammer/shortuuid"
)

type OrderManager struct {
	FiatThreshold               float64       // USD // 2.0 for Testing, 10 USD for production
	OrderTimeOut                time.Duration // 2 * time.Hour
	ExConfirmationThreshold     int           // 10
	WalletConfirmationThreshold int           // 3
	TestingAmount               float64       // 0.00001
	Hestia                      services.HestiaService
	Plutus                      services.PlutusService
}

func (om *OrderManager) GetOrderMap() map[string][]hestia.AdrestiaOrder {
	var adrestiaOrders = adrestia.OrderParams{
		IncludeComplete: true,
		AddedSince:      0,
	}
	orders, err := om.Hestia.GetAllOrders(adrestiaOrders)
	if err != nil {
		fmt.Println(err)
	}
	var mappedOrders = make(map[string][]hestia.AdrestiaOrder)
	for _, order := range orders {
		mappedOrders[hestia.AdrestiaStatusStr[order.Status]] = append(mappedOrders[hestia.AdrestiaStatusStr[order.Status]], order)
	}
	return mappedOrders
}

func (om *OrderManager) ExchangeCanFulfill(exchange exchanges.IExchange, coin string, amount float64) (bool, error) {
	var exAmount, err = exchange.GetBalances()
	if err != nil {
		return false, err
	}
	for _, asset := range exAmount {
		if asset.Ticker == coin {
			if asset.ConfirmedBalance > amount {
				return true, nil
			}
			if asset.ConfirmedBalance+asset.UnconfirmedBalance > amount {
				return false, nil
			}
		}
	}
	return false, nil
}

func (om *OrderManager) HandleBalances() {
	/*
		Fetches information about exchanges, their pending orders
	*/
}

func (om *OrderManager) HandleSentOrders(orders []hestia.AdrestiaOrder) {
	ef := new(exchanges.ExchangeFactory)
	for _, order := range orders {
		tx, err := om.Plutus.GetWalletTx(order.FromCoin, order.HETxId)
		if err != nil {
			fmt.Println(err)
		}
		if tx.Confirmations > om.ExConfirmationThreshold {
			coinInfo, _ := coinfactory.GetCoin(order.FromCoin)
			ex, err := ef.GetExchangeByCoin(*coinInfo)
			if err != nil {
				continue
			}
			// TODO CreateOrder Method ex.CreateOrder()
			orderId, err := "bdiwbfdusbfdsfdfsd", nil // CreateOrder()
			fmt.Println(ex.GetName())
			if err != nil {
				color.Error.Tips(fmt.Sprintf("%v", err))
				continue
			}
			order.FirstOrder.OrderId = orderId
			order.Status = hestia.AdrestiaStatusCreated
			// TODO Handle rate variation
			order.Amount = 0.05 // TODO Replace with handled rate variation returned in object from CreateOrder()
			_, err = om.Hestia.UpdateAdrestiaOrder(order)
			if err != nil {
				continue
			}
		}
	}
}

func (om *OrderManager) HandleCreatedOrders(orders []hestia.AdrestiaOrder) {
	ef := new(exchanges.ExchangeFactory)
	for _, order := range orders {
		coinInfo, _ := coinfactory.GetCoin(order.ToCoin)
		ex, err := ef.GetExchangeByCoin(*coinInfo) // ex
		if err != nil {
			return
		}
		orderFulfilled := false
		// TODO ex.getOrderStatus // ex.GetOrderStatus(order.OrderId)
		if orderFulfilled {
			// TODO Replace Amount with Order's amount
			conf, err := ex.Withdraw(*coinInfo, order.WithdrawAddress, 0.0)
			// conf, err := ex.Withdraw(*coinInfo, order.WithdrawAddress, order.Amount)
			if err != nil {
				fmt.Println(err)
				// TODO Bot report
				return
			}
			if conf {
				order.Status = hestia.AdrestiaStatusSecondExchange
				ok, err := om.Hestia.UpdateAdrestiaOrder(order)
				if err != nil {
					continue
				}
				fmt.Println("HandleCreatedOrders Status: ", ok)
			}

		}
	}
}

func (om *OrderManager) HandleWithdrawnOrders(orders []hestia.AdrestiaOrder) {
	for _, order := range orders {
		fmt.Println(order)
		// TODO Create exchange method for tracking order status
	}
}

func (om *OrderManager) GetOutwardOrders(balanced []balance.Balance, testingAmount float64) (superavitOrders []hestia.AdrestiaOrder) {
	for _, bWallet := range balanced {
		btcAddress, err := om.Plutus.GetBtcAddress()
		ef := new(exchanges.ExchangeFactory)
		coinInfo, err := coinfactory.GetCoin(bWallet.Ticker)
		if err != nil {
			fmt.Println(err)
			continue
		}
		ex, err := ef.GetExchangeByCoin(*coinInfo)
		if err != nil {
			color.Error.Tips(fmt.Sprintf("%v", err))
		} else {
			// TODO Send to Exchange
			exAddress, err := ex.GetAddress(*coinInfo)
			if err == nil {
				var txInfo = plutus.SendAddressBodyReq{
					Address: exAddress,
					Coin:    coinInfo.Tag,
					Amount:  testingAmount, // TODO Replace with actual amount
				}
				fmt.Println(txInfo)
				txId := "test txId" // txId, _ := services.WithdrawToAddress(txInfo)
				var order hestia.AdrestiaOrder
				order.Status = hestia.AdrestiaStatusCreated
				order.Amount = bWallet.DiffBTC / bWallet.RateBTC
				order.FirstOrder.OrderId = ""
				order.FromCoin = bWallet.Ticker
				order.ToCoin = "BTC"
				order.WithdrawAddress = btcAddress
				order.Time = time.Now().Unix()
				order.Message = "adrestia outward balancing"
				order.ID = shortuuid.New()
				order.FirstOrder.Exchange, _ = ex.GetName()
				order.FirstExAddress = exAddress
				order.HETxId = txId

				superavitOrders = append(superavitOrders, order)
			}
		}
	}
	return
}

func (om *OrderManager) GetInwardOrders(unbalanced []balance.Balance, testingAmount float64) (deficitOrders []hestia.AdrestiaOrder) {
	for _, uWallet := range unbalanced {
		address, err := om.Plutus.GetAddress(uWallet.Ticker)
		ef := new(exchanges.ExchangeFactory)
		coinInfo, err := coinfactory.GetCoin(uWallet.Ticker)
		if err != nil {
			continue
		}
		ex, err := ef.GetExchangeByCoin(*coinInfo)
		if err != nil {
			color.Error.Tips(fmt.Sprintf("%v", err))
		} else {
			// fmt.Println("ex name: ", ex.GetName())
			exAddress, err := ex.GetAddress(*coinfactory.Coins["BTC"])
			if err == nil {
				var txInfo = plutus.SendAddressBodyReq{
					Address: exAddress,
					Coin:    "BTC",
					Amount:  0.0001,
				}
				fmt.Println(txInfo)
				txId := "test txId" // txId, _ := services.WithdrawToAddress(txInfo)
				// TODO Send to Exchange
				var order hestia.AdrestiaOrder
				order.Status = hestia.AdrestiaStatusCreated
				order.Amount = testingAmount // TODO Replace with uWallet.DiffBTC
				order.HETxId = ""
				order.FromCoin = "BTC"
				order.ToCoin = uWallet.Ticker
				order.WithdrawAddress = address
				order.Time = time.Now().Unix()
				order.Message = "adrestia inward balancing"
				order.ID = shortuuid.New()
				order.FirstOrder.Exchange, _ = ex.GetName()
				order.FirstExAddress = exAddress
				order.EHTxId = txId

				deficitOrders = append(deficitOrders, order)
			} else {
				fmt.Println("error ex factory: ", err)
			}
		}
	}
	return
}
