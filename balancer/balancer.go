package main

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/adrestia-go/utils"
	"github.com/grupokindynos/common/obol"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	fmt.Println("OBOL URL", os.Getenv("OBOL_URL"))
	hestiaService := services.HestiaRequests{}
	obolService := obol.ObolRequest{ObolURL: os.Getenv("OBOL_URL")}
	plutusService := services.PlutusRequests{Obol: &obolService}
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
	var balances = plutusService.GetWalletBalances(confHestia)        // Gets balance from Hot Wallets
	 // Firebase Wallet Configuration
	if err != nil {
		log.Fatalln(err)
	}
	availableWallets, _ := utils.NormalizeWallets(balances, confHestia) // Verifies wallets in firebase are the same as in plutus and creates a map

	fmt.Println("Available Wallets", availableWallets)

	balanced, unbalanced := utils.SortBalances(availableWallets)

	fmt.Println(balanced, unbalanced)
	/* var superavitOrders = om.GetOutwardOrders(balanced, testingAmount)
	var deficitOrders = om.GetInwardOrders(unbalanced, testingAmount)

	log.Println(superavitOrders)
	log.Println(deficitOrders) */

	// Stores orders in Firestore for further processing
	//utils.StoreOrders(superavitOrders)
	//utils.StoreOrders(deficitOrders)
}
