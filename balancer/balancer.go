package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/adrestia-go/utils"
	cf "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	cutils "github.com/grupokindynos/common/utils"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

func getExchangeOrder(exchange exchanges.IExchange, fromCoin string, toCoin string) (hestia.ExchangeOrder, error) {
	var exchangeOrder hestia.ExchangeOrder
	var err error
	exchangeOrder.Exchange, err = exchange.GetName()
	if err != nil {
		fmt.Println(err)
		return exchangeOrder, err
	}

	orderSide, err := exchange.GetPair(fromCoin, toCoin)
	if err != nil {
		fmt.Println(err)
		return exchangeOrder, err
	}
	exchangeOrder.Symbol = orderSide.Book
	exchangeOrder.Side = orderSide.Type
	exchangeOrder.ReceivedCurrency = toCoin
	exchangeOrder.SoldCurrency = fromCoin

	return exchangeOrder, nil
}

func main() {
	fmt.Println("OBOL URL", os.Getenv("OBOL_URL"))
	hestiaService := services.HestiaRequests{}
	obolService := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	plutusService := services.PlutusRequests{Obol: &obolService}
	exFactory := exchanges.NewExchangeFactory(exchanges.Params{Obol: &obolService})
	color.Info.Tips("Program Started")
	/*
		Process Description
		Check for wallets with superavits, send remaining to exchange conversion to bTC and then send to HW.
		Use exceeding balance in HW (or a new bTC WALLET that solely fits this purpose) to balance other wallets
		in exchanges (should convert and withdraw to an address stored in Firestore).
	*/
	// TODO Disable and Enable Shift at start and re-enable ending of the process

	// TODO This should be the last process, accounting for moved orders
	confHestia, err := hestiaService.GetAdrestiaCoins()
	fmt.Println(confHestia)
	var balances = plutusService.GetWalletBalances(confHestia) // Gets balance from Hot Wallets
	// Firebase Wallet Configuration
	if err != nil {
		log.Fatalln(err)
	}
	availableWallets, _ := utils.NormalizeWallets(balances, confHestia) // Verifies wallets in firebase are the same as in plutus and creates a map

	fmt.Println("Available Wallets", availableWallets)

	balanced, unbalanced := utils.SortBalances(availableWallets)

	fmt.Println(balanced, unbalanced)
	txs := utils.BalanceHW(balanced, unbalanced)

	for _, tx := range txs {
		var firstAddress string
		var secondAddress string
		var firstExchangeOrder hestia.ExchangeOrder
		var secondExchangeOrder hestia.ExchangeOrder
		dualExchange := false

		coin, err := cf.GetCoin(tx.FromCoin)
		if err != nil {
			fmt.Println(err)
			continue
		}
		exchange, err := exFactory.GetExchangeByCoin(*coin)
		if err != nil {
			fmt.Println(err)
			continue
		}
		firstAddress, err = exchange.GetAddress(*coin)
		if err != nil {
			fmt.Println(err)
			continue
		}

		firstExchangeOrder, err = getExchangeOrder(exchange, tx.FromCoin, "BTC")
		if err != nil {
			fmt.Println(err)
			continue
		}

		if tx.ToCoin != "BTC" {
			coin, err = cf.GetCoin(tx.ToCoin)
			if err != nil {
				fmt.Println(err)
				continue
			}
			exchange, err = exFactory.GetExchangeByCoin(*coin)
			if err != nil {
				fmt.Println(err)
				continue
			}
			secondAddress, err = exchange.GetAddress(*coin)
			if err != nil {
				fmt.Println(err)
				continue
			}
			secondExchangeOrder, err = getExchangeOrder(exchange, "BTC", tx.ToCoin)
			if err != nil {
				fmt.Println(err)
				continue
			}
			dualExchange = true
		}

		hwAddress, err := plutusService.GetAddress(tx.ToCoin)
		if err != nil {
			fmt.Println(err)
			continue
		}

		order := hestia.AdrestiaOrder{
			ID:              cutils.RandomString(),
			DualExchange:    dualExchange,
			CreatedTime:     time.Now().Unix(),
			Status:          hestia.AdrestiaStatusCreated,
			Amount:          tx.Amount,
			BtcRate:         tx.BtcRate,
			FromCoin:        tx.FromCoin,
			ToCoin:          tx.ToCoin,
			FirstOrder:      firstExchangeOrder,
			FinalOrder:      secondExchangeOrder,
			FirstExAddress:  firstAddress,
			SecondExAddress: secondAddress,
			WithdrawAddress: hwAddress,
		}

		// symbol side exchange name ReferenceCurrency listingCurrency
		_, err = hestiaService.CreateAdrestiaOrder(order)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
