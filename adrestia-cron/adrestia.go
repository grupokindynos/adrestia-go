package main

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/adrestia-cron/models"
	apiServices "github.com/grupokindynos/adrestia-go/api/services"
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
	"sort"
	"time"

	"github.com/grupokindynos/adrestia-go/models/transaction"

	"github.com/grupokindynos/adrestia-go/models/balance"
)

const fiatThreshold = 2.00 // USD // 2.0 for Testing, 10 USD for production
const orderTimeOut = 2 * time.Hour
const exConfirmationThreshold = 10
const walletConfirmationThreshold = 3

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	/*
	Process Description
	Check for wallets with superavits, send remaining to exchange conversion to bTC and then send to HW.
	Use exceedent balance in HW (or a new bTC WALLET that solely fits this purpose) to balance other wallets in exchanges (should convert and withdraw to an address stored in Firestore).
	 */
	om := new(models.OrderManager)
	orders := om.GetOrderMap()

	// First case: verify sent orders
	sentOrders := orders[hestia.AdrestiaStatusStr[hestia.AdrestiaStatusSentAmount]]
	fmt.Print(sentOrders)
	for _, order := range sentOrders {
		fmt.Print(order)
	}


	// TODO Disable and Enable Shift at star nd ending of the process
	color.Info.Tips("Program Started")
	var balances = services.GetWalletBalances()				// Gets balance from Hot Wallets
	confHestia, err := services.GetCoinConfiguration()		// Firebase Wallet ConfiguratioN

	if err != nil {
		log.Fatalln(err)
	}
	availableWallets, errorWallets := NormalizeWallets(balances, confHestia) // Verifies wallets in firebase are the same as in plutus and creates a map

	fmt.Println("Balancing these Wallets: ", availableWallets)
	fmt.Println("Errors on these Wallets: ", errorWallets)

	balanced, unbalanced := SortBalances(availableWallets)

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
			if err != nil {
				var txInfo = plutus.SendAddressBodyReq{
					Address: exAddress,
					Coin:    coinInfo.Tag,
					Amount:  0.0001,
				}
				fmt.Println(txInfo)
				txId := "test txId"// txId, _ := services.WithdrawToAddress(txInfo)
				var order hestia.AdrestiaOrder
				order.Status = hestia.AdrestiaStatusStr[hestia.AdrestiaStatusSentAmount]
				order.Amount = bWallet.DiffBTC / bWallet.RateBTC
				order.OrderId = "get order from method"
				order.FromCoin = bWallet.Ticker
				order.ToCoin = "BTC"
				order.WithdrawAddress = btcAddress
				order.Time = time.Now().Unix()
				order.Message = "Adrestia outward balancing"
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
		ex, err := ef.GetExchangeByCoin(*coinInfo)
		if err != nil {
			color.Error.Tips(fmt.Sprintf("%v", err))
		} else {
			// TODO Order posting
			var order hestia.AdrestiaOrder
			order.Status = hestia.AdrestiaStatusStr[hestia.AdrestiaStatusCreated]
			order.Amount = uWallet.DiffBTC / uWallet.RateBTC
			order.OrderId = "get order from method"
			order.FromCoin = "btc"
			order.ToCoin = uWallet.Ticker
			order.WithdrawAddress = address
			order.Time = time.Now().Unix()
			order.Message = "Adrestia inward balancing"
			order.ID = shortuuid.New()
			order.Exchange, _ = ex.GetName()

			deficitOrders = append(deficitOrders, order)
		}
	}
	log.Println(superavitOrders)
	log.Println(deficitOrders)
	StoreOrders(superavitOrders)
	StoreOrders(deficitOrders)
	panic("stop!!")
	status, amount := DetermineBalanceability(balanced, unbalanced)
	log.Println(fmt.Sprintf("Wallets Balanceability Status: %t\nAmount (+/-): %.8f", status, amount))
	if status {
		log.Println("balancing...")
		// Calculates Balancing txes
		sendToExchanges := BalanceHW(balanced, unbalanced)

		fmt.Println(sendToExchanges)
		fmt.Println("Found Txes: ", sendToExchanges)

		orders, _ := SendToExchanges(sendToExchanges)

		StoreOrders(orders)
	} else {
		log.Println(fmt.Sprintf("can not balance by missing %.8f BTC in BTC HotWallet", amount))
	}

	panic("stop!")
	// Evaluate wallets with exceeding amount
	for i, w := range availableWallets {
		fmt.Println("Retrieving for ", i, " ", w)
		if w.FirebaseConf.Balances.HotWallet < w.HotWalletBalance.GetTotalBalance() {
			tx := new(transaction.PTx)
			tx.FromCoin = w.HotWalletBalance.Ticker
			tx.ToCoin = w.HotWalletBalance.Ticker
			tx.Amount = w.FirebaseConf.Balances.HotWallet - w.HotWalletBalance.GetTotalBalance()
			tx.Rate = 1.0
			//sendToExchanges = append(sendToExchanges, *tx)
		}
	}
	/*
	// Send remaining amount to exchanges using plutus
	for _, tx := range sendToExchanges{
		fmt.Println("------------ TX-----------")
		fmt.Println(tx)
		coinInfo, err := coinfactory.GetCoin(tx.FromCoin)
		if err != nil {
			color.Error.Tips("%s", err)
			continue
		}
		// ex, err := ef.GetExchangeByName(coinInfo.Rates.Exchange)
		if err != nil {
			color.Error.Tips("%s", err)
			continue
		}
		// add, err :=ex.GetAddress(*coinInfo)
		if err != nil {
			color.Error.Tips("%s", err)
			continue
		}
		// color.Info.Tips("Sending %.8f %s to its exchange at %s", tx.Amount, tx.FromCoin, add)
	}
	// fmt.Println(conf)

	 */
	/* var balanced, unbalanced = SortBalances(balances, conf)

	isBalanceable, diff := DetermineBalanceability(balanced, unbalanced)
	if isBalanceable {
		fmt.Printf("Wallet is balanceable by %.8f\n", diff)
		BalanceHW(balanced, unbalanced) // Balances HW
	} else {
		fmt.Printf("Wallet is not balanceable by %.8f\n", diff)
		BalanceHW(balanced, unbalanced)
		/*
			TODO Handle buy and sell requests on Adrestia as well as proper retrial
			on condition fulfillments

	}*/
}

// Retrieves minimum set balance configuration from test data
func GetFBConfiguration(file string) map[string]balance.Balance {
	jsonFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened conf.json")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var conf map[string]balance.Balance
	err = json.Unmarshal(byteValue, &conf)
	if err != nil {
		log.Println(err)
	}
	return conf
}

// Sorts Balances given their diff, so that topped wallets are used to fill the missing ones
func SortBalances(data map[string]balance.WalletInfoWrapper) ([]balance.Balance, []balance.Balance) {
	var balancedWallets []balance.Balance
	var unbalancedWallets []balance.Balance

	for _, obj := range data {
		x := obj.HotWalletBalance // Curreny Wallet Info Wrapper
		x.GetDiff()
		if x.IsBalanced {
			balancedWallets = append(balancedWallets, x)
		} else {
			unbalancedWallets = append(unbalancedWallets, x)
		}
	}

	sort.Sort(balance.ByDiffInverse(balancedWallets))
	sort.Sort(balance.ByDiff(unbalancedWallets))
	for _, wallet := range unbalancedWallets {
		fmt.Printf("%s has a deficit of %.8f BTC\n", wallet.Ticker, wallet.DiffBTC)
	}
	for _, wallet := range balancedWallets {
		fmt.Printf("%s has a superavit of %.8f BTC\n", wallet.Ticker, wallet.DiffBTC)
	}
	return balancedWallets, unbalancedWallets
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

func NormalizeWallets(balances []balance.Balance, hestiaConf []hestia.Coin) (map[string]balance.WalletInfoWrapper, []string){
	/*
		This function normalizes the wallets that were detected in Plutus and those with configuration in Hestia.
		Returns a map of the coins' ticker as key containing a wrapper with both the actual balance of the wallet and
		its firebase configuration.
	*/
	var mapBalances = make(map[string]balance.Balance)
	var mapConf =  make(map[string]hestia.Coin)
	var missingCoins []string
	var availableCoins = make(map[string]balance.WalletInfoWrapper)

	for _, b := range balances {
		mapBalances[b.Ticker] = b
	}
	for _, c := range hestiaConf {
		mapConf[c.Ticker] = c
	}

	for _, elem := range mapBalances {
		_, ok := mapConf[elem.Ticker]
		if !ok {
			missingCoins = append(missingCoins, elem.Ticker)
		} else {
			/*
				If the current coin is present in both the coinConfig and the acquired Balance maps,
			 	the proceed with the wrapper creation that will handle the balancing of the coins.
			 */
			fmt.Println(elem.Ticker, "\n", mapConf[elem.Ticker].Balances.HotWallet)
			elem.SetTarget(mapConf[elem.Ticker].Balances.HotWallet) // Final attribute for Balance class, represents the target amount in the base currency that should be present
			if elem.Target > 0.0 {
				availableCoins[elem.Ticker] = balance.WalletInfoWrapper{
					HotWalletBalance: elem,
					FirebaseConf:     mapConf[elem.Ticker],
				}
			}
		}
	}
	return availableCoins, missingCoins
}

func SendToExchanges(sendToExchanges []transaction.PTx) (adrestiaOrders []hestia.AdrestiaOrder, err error){
	// Order Creation
	ef := new(apiServices.ExchangeFactory)

	for _, tx := range sendToExchanges {
		var order hestia.AdrestiaOrder
		coinInfo, err := coinfactory.GetCoin(tx.ToCoin)
		if err != nil {
			fmt.Println(err)
		} else {
			ex, err := ef.GetExchangeByCoin(*coinInfo)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("debugging: ", tx.Amount)
				address, _ := ex.GetAddress(*coinInfo)
				fmt.Print(address)
				order.ToCoin = tx.ToCoin
				order.FromCoin = tx.FromCoin
				order.Amount = tx.Amount
				order.Status = hestia.AdrestiaStatusStr[hestia.AdrestiaStatusCompleted]
				order.Exchange, _ = ex.GetName()
				order.ID = shortuuid.New()
				order.Message = "testing adrestia 15-Nov"
				order.Time = time.Now().Unix()
				order.WithdrawAddress = address
				order.OrderId = "pending match with order id"

				adrestiaOrders = append(adrestiaOrders, order)
				res, err := services.CreateAdrestiaOrder(order)
				if err != nil {
					fmt.Println("Error posting order to Hestia", err)
				}
				fmt.Println(res)
			}
		}
		fmt.Println(adrestiaOrders)
	}
	return adrestiaOrders, nil
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