/*
	Process Description
	Check for wallets with superavits, send remaining to exchange conversion to bTC and then send to HW.
	Use exceeding balance in HW (or a new bTC WALLET that solely fits this purpose) to balance other wallets
	in exchanges (should convert and withdraw to an address stored in Firestore).
*/

package main

import (
	"flag"
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
	exchangeOrder.ReceivedCurrency = orderSide.ReceivedCurrency
	exchangeOrder.SoldCurrency = orderSide.SoldCurrency

	return exchangeOrder, nil
}

func getOrderInfo(exFact exchanges.ExchangeFactory, exchangeCoin string, addressCoin string, orderFromCoin string, orderToCoin string) (string, hestia.ExchangeOrder, error) {
	coin, err := cf.GetCoin(exchangeCoin)
	if err != nil {
		fmt.Println(err)
		return "", hestia.ExchangeOrder{}, err
	}
	addrCoin, err := cf.GetCoin(addressCoin)
	if err != nil {
		fmt.Println(err)
		return "", hestia.ExchangeOrder{}, err
	}
	exchange, err := exFact.GetExchangeByCoin(*coin)
	if err != nil {
		fmt.Println(err)
		return "", hestia.ExchangeOrder{}, err
	}
	address, err := exchange.GetAddress(*addrCoin)
	if err != nil {
		fmt.Println("116 - ", err)
		return "", hestia.ExchangeOrder{}, err
	}

	exchangeOrder, err := getExchangeOrder(exchange, orderFromCoin, orderToCoin)
	if err != nil {
		fmt.Println("122 - ", err)
		return "", hestia.ExchangeOrder{}, err
	}

	return address, exchangeOrder, nil
}

var (
	hestiaEnv string
	plutusEnv string
)

func main() {
	// Read input flag
	localRun := flag.Bool("local", false, "set this flag to run adrestia with local db")
	flag.Parse()

	// If flag was set, change the hestia request url to be local
	if *localRun {
		hestiaEnv = "HESTIA_LOCAL_URL"
		plutusEnv = "PLUTUS_LOCAL_URL"
	} else {
		hestiaEnv = "HESTIA_PRODUCTION_URL"
		plutusEnv = "PLUTUS_PRODUCTION_URL"
	}

	hestiaService := services.HestiaRequests{HestiaURL: os.Getenv(hestiaEnv)}
	obolService := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	plutusService := services.PlutusRequests{Obol: &obolService, PlutusURL: os.Getenv(plutusEnv)}
	exFactory := exchanges.NewExchangeFactory(exchanges.Params{Obol: &obolService})

	color.Info.Tips("Program Started")

	confHestia, err := hestiaService.GetAdrestiaCoins()
	//fmt.Println(confHestia)
	var balances = plutusService.GetWalletBalances(confHestia) // Gets balance from Hot Wallets
	// Firebase Wallet Configuration
	if err != nil {
		log.Fatalln(err)
	}
	availableWallets, _ := utils.NormalizeWallets(balances, confHestia) // Verifies wallets in firebase are the same as in plutus and creates a map

	fmt.Println("Available Wallets", availableWallets)

	balanced, unbalanced := utils.SortBalances(availableWallets)
	isBalanceable, diff := utils.DetermineBalanceability(balanced, unbalanced)
	if !isBalanceable {
		fmt.Println("HW cannot be balanced, deficit greater that superavit by", -diff)
		return
	}

	fmt.Println("Finished sorting")
	fmt.Println(balanced, unbalanced)
	txs := utils.BalanceHW(balanced, unbalanced)

	for _, tx := range txs {
		var firstAddress string
		var secondAddress string
		var firstExchangeOrder hestia.ExchangeOrder
		var secondExchangeOrder hestia.ExchangeOrder
		dualExchange := false

		if tx.FromCoin != "BTC" {
			firstAddress, firstExchangeOrder, err = getOrderInfo(*exFactory, tx.FromCoin, tx.FromCoin, tx.FromCoin, "BTC")
			if err != nil {
				fmt.Println("122 - ", err)
				continue
			}
		} else {
			firstAddress, firstExchangeOrder, err = getOrderInfo(*exFactory, tx.ToCoin, tx.FromCoin, tx.FromCoin, tx.ToCoin)
			if err != nil {
				fmt.Println("122 - ", err)
				continue
			}
		}

		if tx.ToCoin != "BTC" && tx.FromCoin != "BTC" {
			secondAddress, secondExchangeOrder, err = getOrderInfo(*exFactory, tx.ToCoin, "BTC", "BTC", tx.ToCoin)
			if err != nil {
				fmt.Println("144 - ", err)
				continue
			}
			dualExchange = true
		}

		hwAddress, err := plutusService.GetAddress(tx.ToCoin)
		if err != nil {
			fmt.Println("152 - ", err)
			continue
		}
		log.Println("Finish Get address")

		order := hestia.AdrestiaOrder{
			ID:              tx.FromCoin + tx.ToCoin + cutils.RandomString(),
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
		log.Println("Finish order")
		_, err = hestiaService.CreateAdrestiaOrder(order)
		if err != nil {
			fmt.Println("174 - ", err)
			continue
		}
	}
}
