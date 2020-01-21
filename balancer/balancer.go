package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/exchanges"
	"github.com/grupokindynos/adrestia-go/models/exchange_models"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/adrestia-go/utils"
	cf "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/utils"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	fmt.Println("OBOL URL", os.Getenv("OBOL_URL"))
	hestiaService := services.HestiaRequests{}
	obolService := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	plutusService := services.PlutusRequests{Obol: &obolService}
	exFactory := exchanges.NewExchangeFactory(exchange_models.Params{Obol: &obolService})
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
			dualExchange = true
		}
		hwAddress, err := plutusService.GetAddress(tx.ToCoin)
		if err != nil {
			fmt.Println(err)
			continue
		}

		order := hestia.AdrestiaOrder{
			ID:              utils.RandomString(),
			DualExchange:    dualExchange,
			Time:            time.Now().Unix(),
			Status:          hestia.AdrestiaStatusCreated,
			Amount:          tx.Amount,
			BtcRate:         tx.BtcRate,
			FromCoin:        tx.FromCoin,
			ToCoin:          tx.ToCoin,
			FirstExAddress:  firstAddress,
			SecondExAddress: secondAddress,
			WithdrawAddress: hwAddress,
		}

		_, err = hestiaService.CreateAdrestiaOrder(order)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
