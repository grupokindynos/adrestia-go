package main

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/color"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"

	"github.com/grupokindynos/adrestia-go/models/transaction"
	CoinFactory "github.com/grupokindynos/common/coin-factory"
	obol "github.com/grupokindynos/common/obol"

	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/services"
)

var printDebugInfo = true
const plutusUrl = "https://plutus-wallets.herokuapp.com"

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	color.Info.Tips("Program Started")
	// coins := []string{ "POLIS", "DASH" }

	// Gets balance from Hot Wallets
	var balances = GetWalletBalances()
	if printDebugInfo {
		color.Info.Tips("\t\tAvailable Coins")
		for i, _ := range balances {
			color.Info.Tips(fmt.Sprintf("Wallet has %.8f %s", balances[i].Balance, balances[i].Ticker))
		}
	}
	// Firebase Wallet Configuration
	var conf = GetFBConfiguration()
	// fmt.Println(conf)
	var balanced, unbalanced = SortBalances(balances, conf)

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
		*/
	}

}

func GetWalletBalances() []balance.Balance {
	fmt.Println("\tRetrieving Wallet Balances...")

	var rawBalances []balance.Balance

	// TODO Fix data retrieval from Plutus
	// availableCoins := CoinFactory.Coins
	/*fmt.Println("1. ", os.Getenv("PLUTUS_URL"))
	fmt.Println("2. ", os.Getenv("PLUTUS_AUTH_USERNAME"))
	fmt.Println("3. ", os.Getenv("PLUTUS_AUTH_PASSWORD"))
	fmt.Println("4. ", os.Getenv("PLUTUS_PUBLIC_KEY"))
	fmt.Println("5. ", os.Getenv("TYCHE_PUBLIC_KEY"))
	fmt.Println("6. ", os.Getenv("MASTER_PASSWORD"))
	for _, coin := range availableCoins {
		res, err := plutus.GetWalletAddress(os.Getenv("PLUTUS_URL"), "BTC", os.Getenv("TYCHE_PRIV_KEY"), "tyche", os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("PLUTUS_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
		//res, err := plutus.GetWalletBalance(os.Getenv("PLUTUS_URL"), strings.ToLower(coin.Tag), os.Getenv("TYCHE_PUBLIC_KEY"), "tyche", os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("PLUTUS_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
		fmt.Println("debug: ", res, coin.Tag)
		if err != nil {
			fmt.Println("Plutus Service Error: ", err)
		}
		// Create Balance Object
		/*ar b = balance.Balance{}
		b.Balance = res.Confirmed
		b.Ticker = coin.Tag

		rawBalances = append(rawBalances, b)

	}*/
	rawBalances, err := loadTestingData()
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("Raw Balances: ", rawBalances)
	fmt.Println("Finished Retrieving Balances")

	var updatedBalances []balance.Balance
	fmt.Println("\tRetrieving Wallet Rates...")
	for _, coin := range rawBalances {
		// fmt.Println(coin)
		var currentBalance = coin
		rate, err := obol.GetCoin2CoinRates(currentBalance.Ticker, "btc")
		if err != nil{
			color.Error.Tips(fmt.Sprintf("Rate failed for %s. Error: %s", coin.Ticker, err))
		} else {
			color.Info.Tips(fmt.Sprintf("Rate retrieved for %s.", coin.Ticker))
			currentBalance.RateBTC = rate
			updatedBalances = append(updatedBalances, currentBalance)
		}

	}
	return updatedBalances
}

// Retrieves minimum set balance configuration from Firebase conf
func GetFBConfiguration() map[string]balance.Balance {

	fireBase := services.InitFirebase()
	var conf balance.MinBalanceConfResponse
	conf, err := fireBase.GetConf()
	if err != nil {
		log.Fatal("Configuration not found")
	}
	// fmt.Println("Retrieved config: ", conf)

	var firebaseConfBalances = conf.ToMap()

	return firebaseConfBalances

}

func SortBalances(inputBalances []balance.Balance, conf map[string]balance.Balance) ([]balance.Balance, []balance.Balance) {
	// Sorts Balances

	var balancedWallets []balance.Balance
	var unbalancedWallets []balance.Balance

	for _, obj := range inputBalances {
		obj.GetDiff(conf[obj.Ticker].Balance)
		if obj.IsBalanced {
			balancedWallets = append(balancedWallets, obj)
		} else {
			unbalancedWallets = append(unbalancedWallets, obj)
		}
	}

	sort.Sort(balance.ByDiffInverse(balancedWallets))
	sort.Sort(balance.ByDiff(unbalancedWallets))

	fmt.Println("Info Sorting")
	fmt.Println()
	for _, wallet := range unbalancedWallets {
		fmt.Printf("%s has a deficit of %.8f BTC\n", wallet.Ticker, wallet.DiffBTC)
	}
	for _, wallet := range balancedWallets {
		fmt.Printf("%s has a superavit of %.8f BTC\n", wallet.Ticker, wallet.DiffBTC)
	}

	return balancedWallets, unbalancedWallets
}

func DetermineBalanceability(balanced []balance.Balance, unbalanced []balance.Balance) (bool, float64) {
	superavit := 0.0 // Excedent in balanced wallets
	deficit := 0.0   // Missing amount in unbalanced wallets

	for _, wallet := range balanced {
		superavit += wallet.DiffBTC
	}
	for _, wallet := range unbalanced {
		deficit += wallet.DiffBTC
	}
	fmt.Println("Superavit: ", superavit)
	fmt.Println("Deficit: ", deficit)
	return superavit > math.Abs(deficit), superavit - math.Abs(deficit)
}

// Actual balancing action
func BalanceHW(balanced []balance.Balance, unbalanced []balance.Balance) []transaction.PTx {
	var pendingTransactions []transaction.PTx
	bIndex := 0
	for _, wallet := range unbalanced {
		// fmt.Println(i, " ", wallet)
		coinData, _ := CoinFactory.GetCoin(wallet.Ticker)

		color.Info.Tips(fmt.Sprintf("The exchange for %s is %s\n", wallet.Ticker, coinData.Rates.Exchange))
		// TODO Optimize sending TXs for the same coin (instead of making 5 dash transactions, make one)
		if wallet.DiffBTC < balanced[bIndex].DiffBTC {
			var newTx = transaction.PTx{
				ToCoin:   wallet.Ticker,
				FromCoin: balanced[bIndex].Ticker,
				Amount:   math.Abs(wallet.DiffBTC),
			}
			pendingTransactions = append(pendingTransactions, newTx)
		}
	}
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
