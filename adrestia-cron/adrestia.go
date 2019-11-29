package main

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/adrestia-cron/models"
	"github.com/grupokindynos/adrestia-go/adrestia-cron/utils"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"github.com/joho/godotenv"
	"log"
	"time"
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
	// TODO Disable and Enable Shift at star nd ending of the process
	color.Info.Tips("Program Started")
	/*
		Process Description
		Check for wallets with superavits, send remaining to exchange conversion to bTC and then send to HW.
		Use exceeding balance in HW (or a new bTC WALLET that solely fits this purpose) to balance other wallets
		in exchanges (should convert and withdraw to an address stored in Firestore).
	 */
	om := models.NewOrderManager(fiatThreshold, orderTimeOut, exConfirmationThreshold, walletConfirmationThreshold, testingAmount)
	orders := om.GetOrderMap()

	// First case: verify sent orders
	sentOrders := orders[hestia.AdrestiaStatusStr[hestia.AdrestiaStatusSentAmount]]
	createdOrders := orders[hestia.AdrestiaStatusStr[hestia.AdrestiaStatusCreated]]
	partiallyFulfilledOrders := orders[hestia.AdrestiaStatusStr[hestia.AdrestiaStatusPartiallyFulfilled]]
	awaitingWithdrawOrders := orders[hestia.AdrestiaStatusStr[hestia.AdrestiaStatusPendingWidthdrawal]]
	completedOrders := orders[hestia.AdrestiaStatusStr[hestia.AdrestiaStatusCompleted]]

	fmt.Print("Sent Orders: ", sentOrders)
	fmt.Println("Created Orders: ", createdOrders)

	fmt.Println(partiallyFulfilledOrders)
	fmt.Println(awaitingWithdrawOrders)
	fmt.Println(completedOrders)

	// TODO This should be the last process, accounting for moved orders
	var balances = services.GetWalletBalances()				// Gets balance from Hot Wallets
	confHestia, err := services.GetCoinConfiguration()		// Firebase Wallet Configuration
	if err != nil {
		log.Fatalln(err)
	}
	availableWallets, _ := utils.NormalizeWallets(balances, confHestia) // Verifies wallets in firebase are the same as in plutus and creates a map
	balanced, unbalanced := utils.SortBalances(availableWallets)

	var superavitOrders = models.GetOutwardOrders(balanced, testingAmount)
	var deficitOrders = models.GetInwardOrders(unbalanced, testingAmount)

	log.Println(superavitOrders)
	log.Println(deficitOrders)

	// Stores orders in Firestore for further processing
	utils.StoreOrders(superavitOrders)
	utils.StoreOrders(deficitOrders)
}
