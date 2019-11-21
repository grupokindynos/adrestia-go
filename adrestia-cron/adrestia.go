package main

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/adrestia-cron/models"
	"github.com/grupokindynos/adrestia-go/adrestia-cron/utils"
	apiServices "github.com/grupokindynos/adrestia-go/api/services"
	"github.com/grupokindynos/adrestia-go/services"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/plutus"
	"github.com/joho/godotenv"
	"github.com/lithammer/shortuuid"
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
	fmt.Print(sentOrders)

	var balances = services.GetWalletBalances()				// Gets balance from Hot Wallets
	confHestia, err := services.GetCoinConfiguration()		// Firebase Wallet Configuration
	if err != nil {
		log.Fatalln(err)
	}
	availableWallets, errorWallets := utils.NormalizeWallets(balances, confHestia) // Verifies wallets in firebase are the same as in plutus and creates a map

	fmt.Println("Balancing these Wallets: ", availableWallets)
	fmt.Println("Errors on these Wallets: ", errorWallets)

	balanced, unbalanced := utils.SortBalances(availableWallets)

	var superavitOrders []hestia.AdrestiaOrder
	var deficitOrders []hestia.AdrestiaOrder
	for _, bWallet := range balanced {
		btcAddress, err := services.GetBtcAddress()
		ef := new(apiServices.ExchangeFactory)
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

	for _, uWallet := range unbalanced {
		address, err := services.GetAddress(uWallet.Ticker)
		ef := new(apiServices.ExchangeFactory)
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
	log.Println(superavitOrders)
	log.Println(deficitOrders)
	utils.StoreOrders(superavitOrders)
	utils.StoreOrders(deficitOrders)
}
