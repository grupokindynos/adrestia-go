package balancer

import (
	"errors"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/utils"
	"log"
	"math"
	"sort"
	"time"
)

type Balancer struct {
	Hestia services.HestiaService
	Plutus services.PlutusService
	Obol   obol.ObolService
}

type balance struct {
	coin string
	difference float64
}

var (
	balancerId string
	balancerOrders []hestia.BalancerOrder
	exFactory *exchanges.ExchangeFactory
)

func (b *Balancer) Start(id string) error {
	coinsProp, err := b.Hestia.GetAdrestiaCoins()
	if err != nil {
		log.Println("Unable to get adrestia coins " + err.Error())
		return err
	}
	exFactory = exchanges.NewExchangeFactory(b.Obol, b.Hestia)
	balancerId = id
	totalHwStock := 0.0
	totalCoins := 0
	stockByCoin := make(map[string]float64)
	for _, coin := range coinsProp{
		balance, err := b.Plutus.GetWalletBalance(coin.Ticker)
		if err != nil {
			log.Println("Unable to get balance " + err.Error())
			return err
		}
		totalCoins += coin.Adrestia.CoinUsage
		totalHwStock += balance.Confirmed
		stockByCoin[coin.Ticker] = balance.Confirmed
	}

	var balances []balance
	for _, coin := range coinsProp {
		stockExpected := math.Floor(totalHwStock * float64(coin.Adrestia.CoinUsage) / float64(totalCoins)) // floor to give a threshold of error for precision problems
		balances = append(balances, balance{coin:coin.Ticker, difference: stockByCoin[coin.Ticker] - stockExpected})
	}
	sort.Slice(balances, func(i, j int) bool{
		return balances[i].difference > balances[j].difference
	})

	index := 0
	for i := 1; i < len(balances); i++ {
		if  i == index {
			return errors.New("impossible to balance hot wallet")
		}
		bi := balances[i]
		//if bi.difference > -50 { // threshold for missing not too much balance
		//	continue
		//}
		if balances[index].difference >= -bi.difference {
			balances[index].difference += bi.difference
			err := b.createBalancerOrder(balances[index].coin, bi.coin, -bi.difference)
			if err != nil {
				return errors.New("Error while creating order "+ err.Error())
			}
			bi.difference = 0
		} else if balances[index].difference > 50 {
			bi.difference += balances[index].difference
			err := b.createBalancerOrder(balances[index].coin, bi.coin, balances[index].difference)
			if err != nil {
				return errors.New("Error while creating order " + err.Error())
			}
			balances[index].difference = 0
		} else {
			index++
			i--
		}
	}

	for _, balancerOrder := range balancerOrders {
		_, err := b.Hestia.CreateBalancerOrder(balancerOrder)
		if err != nil {
			log.Println("Unable to create balancer order " + err.Error())
		}
	}
	return nil
}

func (b *Balancer) createBalancerOrder(fromCoin string, toCoin string, amount float64) error {
	exchange, err := exFactory.GetExchangeByName("binance") // All the stable coin conversions will be performed on binance
	if err != nil {
		return err
	}
	depositAddr, err := exchange.GetAddress(fromCoin)
	if err != nil {
		return err
	}
	withdrawAddr, err := b.Plutus.GetAddress(toCoin)
	if err != nil {
		return err
	}
	dualConversion := false
	trade1 := hestia.Trade{}
	trade2 := hestia.Trade{}
	if fromCoin == "USDT" || toCoin == "USDT" {
		tradeInfo, err := exchange.GetPair(fromCoin, toCoin)
		if err != nil {
			return err
		}
		trade1 = b.createTradeOrder(fromCoin, toCoin, tradeInfo)
	} else {
		dualConversion = true
		tradeInfo1, err := exchange.GetPair(fromCoin, "USDT")
		if err != nil {
			return err
		}
		tradeInfo2, err := exchange.GetPair("USDT", toCoin)
		if err != nil {
			return err
		}
		trade1 = b.createTradeOrder(fromCoin, "USDT", tradeInfo1)
		trade2 = b.createTradeOrder("USDT", toCoin, tradeInfo2)
	}
	
	order := hestia.BalancerOrder {
		Id:              utils.RandomString(),
		BalancerId:      balancerId,
		FromCoin:        fromCoin,
		ToCoin:          toCoin,
		DualConversion:  dualConversion,
		OriginalAmount:  amount,
		FirstTrade:      trade1,
		SecondTrade:     trade2,
		Status:          hestia.BalancerOrderStatusCreated,
		Exchange:        "binance",
		ReceivedAmount:  0,
		DepositTxId:     "",
		WithdrawTxId:    "",
		DepositAddress:  depositAddr,
		WithdrawAddress: withdrawAddr,
		CreatedTime:     time.Now().Unix(),
		FulfilledTime:   0,
	}
	balancerOrders = append(balancerOrders, order)
	return nil
}

func (b *Balancer) createTradeOrder(fromCoin string, toCoin string, info models.TradeInfo) hestia.Trade {
	return hestia.Trade{
		OrderId:        "",
		Amount:         0,
		ReceivedAmount: 0,
		FromCoin:       fromCoin,
		ToCoin:         toCoin,
		Symbol:         info.Book,
		Side:           info.Type,
		CreatedTime:    time.Now().Unix(),
		FulfilledTime:  0,
	}
}
