package main

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/adrestia-cron/models"
	"github.com/grupokindynos/adrestia-go/adrestia-cron/utils"
	apiServices "github.com/grupokindynos/adrestia-go/api/services"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/adrestia-go/services"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/plutus"
	"github.com/joho/godotenv"
	"github.com/lithammer/shortuuid"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"

	"github.com/grupokindynos/adrestia-go/models/balance"
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
	om := new(models.OrderManager)
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
	StoreOrders(superavitOrders)
	StoreOrders(deficitOrders)
}

func DetermineBalanceability(balanced []balance.Balance, unbalanced []balance.Balance) (bool, float64) {
	superavit := 0.0 // Exceeding amount in balanced wallets
	deficit := 0.0   // Missing amount in unbalanced wallets
	totalBtc := 0.0 // total amount in wallets

	for _, wallet := range balanced {
		superavit += math.Abs(wallet.DiffBTC)
		totalBtc += wallet.ConfirmedBalance * wallet.RateBTC // Uses ConfirmedBalance as it is used to balance in a particular adrestia.go run
	}
	for _, wallet := range unbalanced {
		deficit += wallet.DiffBTC
		totalBtc += wallet.GetTotalBalance() * wallet.RateBTC // Uses TotalBalance as it should account for pending Txes expected by coin
	}
	color.Info.Tips(fmt.Sprintf("Total in Wallets: %.8f", totalBtc))
	fmt.Println("Superavit: ", superavit)
	fmt.Println("Deficit: ", deficit)
	return superavit > math.Abs(deficit), superavit - math.Abs(deficit)
}

// Actual balancing action
func BalanceHW(balanced []balance.Balance, unbalanced []balance.Balance) []transaction.PTx {
	var pendingTransactions []transaction.PTx
	i := 0 // Balanced wallet index
	for _, wallet := range unbalanced {
		// log.Println("BalanceHW:: Balancing ", wallet.Ticker)
		filledAmount := 0.0 // Amount that stores current fulfillment of a Balancing Transaction.
		initialDiff := math.Abs(wallet.DiffBTC)
		// TODO add fields for out addresses
		for filledAmount < initialDiff {
			// color.Info.Tips(fmt.Sprintf("BalanceHW::\tUsing %s to balanace %s", balanced[i].Ticker, wallet.Ticker))
			var newTx transaction.PTx
			if balanced[i].DiffBTC < initialDiff - filledAmount {
				newTx.ToCoin = wallet.Ticker
				newTx.FromCoin = balanced[i].Ticker
				newTx.Amount = balanced[i].DiffBTC
				newTx.Rate = balanced[i].RateBTC

				filledAmount += balanced[i].DiffBTC
				balanced[i].DiffBTC = 0.0
				i++
				fmt.Println("Type I tx: ", newTx)
				pendingTransactions = append(pendingTransactions, newTx)
			}else {
				newTx.Amount = initialDiff - filledAmount
				filledAmount += initialDiff - filledAmount
				balanced[i].DiffBTC -= initialDiff - filledAmount
				newTx.ToCoin = wallet.Ticker
				newTx.FromCoin = balanced[i].Ticker
				newTx.Rate = balanced[i].RateBTC
				fmt.Println("Type II tx: ", newTx)
				pendingTransactions = append(pendingTransactions, newTx)
			}
		}
	}
	// TODO Optimization for txes to exchanges
	return pendingTransactions
}

func loadTestingData() ([]balance.Balance, error){
	jsonFile, err := os.Open("test_data/test.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var b []balance.Balance
	err = json.Unmarshal(byteValue, &b)

	if err != nil {
		return b, err
	}
	return b, nil
}

func StoreOrders(orders []hestia.AdrestiaOrder) {
	for _, order := range orders {
		res, err := services.CreateAdrestiaOrder(order)
		if err != nil {
			fmt.Println("error posting order to hestia: ", err)
		} else {
			fmt.Println(res)
		}
	}
}

func HandleBalances() {
	/*
		Fetches information about exchanges, their pending orders
	 */
}

func HandleSentOrders(orders []hestia.AdrestiaOrder) {
	for _, order := range orders {
		tx, err :=services.GetWalletTx(order.FromCoin, order.TxId)
		if err != nil {
			fmt.Println(err)
		}
		if tx.Confirmations > exConfirmationThreshold {
			// TODO Create Order in Exchange and Update Status
			// TODO Handle rate variation
		}
	}
}

func HandleCreatedOrders(orders []hestia.AdrestiaOrder) {
	ef := new(apiServices.ExchangeFactory)
	for _, order := range orders {
		coinInfo, _ := coinfactory.GetCoin(order.ToCoin)
		ex, err := ef.GetExchangeByCoin(*coinInfo)  // ex
		if err != nil {
			return
		}
		orderFulfilled := false
		// TODO ex.getOrderStatus
		if orderFulfilled {
			// TODO Withdraw
			conf, err := ex.Withdraw(*coinInfo, order.WithdrawAddress, 0.0)
			// conf, err := ex.Withdraw(*coinInfo, order.WithdrawAddress, order.Amount)
			if err != nil {
				fmt.Println(err)
				// TODO Bot report
				return
			}
			if conf {
				// TODO Update Status

			}

		}
	}
}

func ChangeOrderStatus(order hestia.AdrestiaOrder, status hestia.AdrestiaStatus) () {
	fallbackStatus := order.Status
	order.Status = hestia.AdrestiaStatusStr[status]
	resp, err := services.UpdateAdrestiaOrder(order)
	// TODO Move in map (if concurrency on maps allows for it)
	if err != nil {
		order.Status = fallbackStatus
		fmt.Println(err)
	} else {
		log.Println(fmt.Sprintf("order %s in %s has been updated to %s\t%s", order.OrderId, order.Exchange, order.Status, resp))
	}


}

func HandleWithdrawnOrders(orders []hestia.AdrestiaOrder) {
	for _, order := range orders {
		fmt.Println(order)
		// TODO Create exchange method for tracking order status
	}
}