package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/handlers"
	"github.com/grupokindynos/adrestia-go/processor"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/adrestia-go/utils"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/joho/godotenv"
)

const fiatThreshold = 2.00 // USD // 2.0 for Testing, 10 USD for production
const orderTimeOut = 2 * time.Hour
const exConfirmationThreshold = 10
const walletConfirmationThreshold = 3
const testingAmount = 0.00001

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	hestiaService := services.HestiaRequests{}
	plutusService := services.PlutusRequests{Obol: &obol.ObolRequest{}}

	proc := processor.Processor{
		Hestia: &hestiaService,
	}

	proc.Start()
	// TODO Disable and Enable Shift at star nd ending of the process
	color.Info.Tips("Program Started")
	/*
		Process Description
		Check for wallets with superavits, send remaining to exchange conversion to bTC and then send to HW.
		Use exceeding balance in HW (or a new bTC WALLET that solely fits this purpose) to balance other wallets
		in exchanges (should convert and withdraw to an address stored in Firestore).
	*/
	om := handlers.OrderManager{
		FiatThreshold:               fiatThreshold,
		OrderTimeOut:                orderTimeOut,
		ExConfirmationThreshold:     exConfirmationThreshold,
		WalletConfirmationThreshold: walletConfirmationThreshold,
		TestingAmount:               testingAmount,
		Hestia:                      &hestiaService,
		Plutus:                      &plutusService,
	}

	orders := om.GetOrderMap()

	// First case: verify sent orders
	createdOrders := orders[hestia.AdrestiaStatusStr[hestia.AdrestiaStatusCreated]]
	firstExchangeOrders := orders[hestia.AdrestiaStatusStr[hestia.AdrestiaStatusFirstExchange]]
	firstConversionOrders := orders[hestia.AdrestiaStatusStr[hestia.AdrestiaStatusFirstConversion]]
	secondExchangeOrders := orders[hestia.AdrestiaStatusStr[hestia.AdrestiaStatusSecondExchange]]
	secondConversionOrders := orders[hestia.AdrestiaStatusStr[hestia.AdrestiaStatusSecondConversion]]
	completedOrders := orders[hestia.AdrestiaStatusStr[hestia.AdrestiaStatusCompleted]]

	fmt.Print("Created Orders: ", createdOrders)
	fmt.Println("Completed Orders: ", completedOrders)

	fmt.Println(firstExchangeOrders)
	fmt.Println(firstConversionOrders)
	fmt.Println(secondExchangeOrders)
	fmt.Println(secondConversionOrders)

	// TODO This should be the last process, accounting for moved orders
	var balances = plutusService.GetWalletBalances()        // Gets balance from Hot Wallets
	confHestia, err := hestiaService.GetCoinConfiguration() // Firebase Wallet Configuration
	if err != nil {
		log.Fatalln(err)
	}
	availableWallets, _ := utils.NormalizeWallets(balances, confHestia) // Verifies wallets in firebase are the same as in plutus and creates a map
	balanced, unbalanced := utils.SortBalances(availableWallets)

	var superavitOrders = om.GetOutwardOrders(balanced, testingAmount)
	var deficitOrders = om.GetInwardOrders(unbalanced, testingAmount)

	log.Println(superavitOrders)
	log.Println(deficitOrders)

	// Stores orders in Firestore for further processing
	//utils.StoreOrders(superavitOrders)
	//utils.StoreOrders(deficitOrders)
}
