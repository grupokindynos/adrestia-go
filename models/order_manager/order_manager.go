package order_manager

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/exchanges"
	services2 "github.com/grupokindynos/adrestia-go/models/services"
	"github.com/grupokindynos/adrestia-go/services"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/plutus"
	"github.com/lithammer/shortuuid"
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
	ef := new(services.ExchangeFactory)
	for _, order := range orders {
		tx, err :=services.GetWalletTx(order.FromCoin, order.TxId)
		if err != nil {
			fmt.Println(err)
		}
		if tx.Confirmations > o.exConfirmationThreshold {
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
			order.OrderId = orderId
			order.Status = hestia.AdrestiaStatusStr[hestia.AdrestiaStatusCreated]
			// TODO Handle rate variation
			order.Amount = 0.05 // TODO Replace with handled rate variation returned in object from CreateOrder()
			_, err = services.UpdateAdrestiaOrder(order)
			if err != nil {
				continue
			}
		}
	}
}

func (o *OrderManager)HandleCreatedOrders(orders []hestia.AdrestiaOrder) {
	ef := new(services.ExchangeFactory)
	for _, order := range orders {
		coinInfo, _ := coinfactory.GetCoin(order.ToCoin)
		ex, err := ef.GetExchangeByCoin(*coinInfo)  // ex
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
				order.Status = hestia.AdrestiaStatusStr[hestia.AdrestiaStatusPendingWidthdrawal]
				ok, err := services.UpdateAdrestiaOrder(order)
				if err != nil {
					continue
				}
				fmt.Println("HandleCreatedOrders Status: ", ok)
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

func GetOutwardOrders(balanced []balance.Balance, testingAmount float64) (superavitOrders []hestia.AdrestiaOrder) {
	for _, bWallet := range balanced {
		btcAddress, err := services.GetBtcAddress()
		ef := new(services.ExchangeFactory)
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
					Amount:  testingAmount,	// TODO Replace with actual amount
				}
				fmt.Println(txInfo)
				txId := "test txId"// txId, _ := services.WithdrawToAddress(txInfo)
				var order hestia.AdrestiaOrder
				order.Status = hestia.AdrestiaStatusStr[hestia.AdrestiaStatusSentAmount]
				order.Amount = bWallet.DiffBTC / bWallet.RateBTC
				order.OrderId = ""
				order.FromCoin = bWallet.Ticker
				order.ToCoin = "BTC"
				order.WithdrawAddress = btcAddress
				order.Time = time.Now().Unix()
				order.Message = "adrestia outward balancing"
				order.ID = shortuuid.New()
				order.Exchange, _ = ex.GetName()
				order.ExchangeAddress = exAddress
				order.TxId = txId

				superavitOrders = append(superavitOrders, order)
			}
		}
	}
	return
}

func GetInwardOrders(unbalanced []balance.Balance, testingAmount float64) (deficitOrders []hestia.AdrestiaOrder) {
	for _, uWallet := range unbalanced {
		address, err := services.GetAddress(uWallet.Ticker)
		ef := new(services.ExchangeFactory)
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
				order.Status = hestia.AdrestiaStatusStr[hestia.AdrestiaStatusSentAmount]
				order.Amount = testingAmount // TODO Replace with uWallet.DiffBTC
				order.OrderId = ""
				order.FromCoin = "BTC"
				order.ToCoin = uWallet.Ticker
				order.WithdrawAddress = address
				order.Time = time.Now().Unix()
				order.Message = "adrestia inward balancing"
				order.ID = shortuuid.New()
				order.Exchange, _ = ex.GetName()
				order.ExchangeAddress = exAddress
				order.TxId = txId

				deficitOrders = append(deficitOrders, order)
			} else {
				fmt.Println("error ex factory: ", err)
			}
		}
	}
	return
}