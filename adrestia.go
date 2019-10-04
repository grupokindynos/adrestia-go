package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"

	"github.com/grupokindynos/adrestia-go/models/transaction"
	CoinFactory "github.com/grupokindynos/common/coin-factory"

	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/services"
	plutus_service "github.com/grupokindynos/tyche/services"
)

const baseUrl string = "https://delphi.polispay.com/api/"

var ratePvdr = services.RateProvider{}
var printDebugInfo = true
var ps = plutus_service.PlutusService{
	// PlutusURL:    "https://localhost:8280",
	PlutusURL:    "https://plutus-wallets.herokuapp.com",
	AuthUsername: os.Getenv("PLUTUS_AUTH_USERNAME"),
	AuthPassword: os.Getenv("PLUTUS_AUTH_PASSWORD"),
}

func main() {
	fmt.Println("Program Started")
	// coins := []string{ "POLIS", "DASH" }

	// Gets balance from Hot Wallets
	var balances = GetWalletBalances()
	if printDebugInfo {
		fmt.Println("\t\tAvailable Coins")
		for i, _ := range balances {
			fmt.Println("\t\t", balances[i].Balance, balances[i].Ticker)
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

	var balances balance.HotWalletBalances
	// response, err := http.Get(baseUrl + "v2/wallets/balances")
	res, err := ps.GetWalletBalance("polis")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Res", res)

	fmt.Println("Finished Retrieving Balances")

	var updatedBalances []balance.Balance
	fmt.Println("\tRetrieving Wallet Rates...")
	for _, coin := range balances.Data {
		// fmt.Println(coin)
		var newBalance = coin
		newBalance.RateBTC = ratePvdr.GetRate(newBalance.Ticker)
		updatedBalances = append(updatedBalances, newBalance)
	}
	// fmt.Println(updatedBalances)
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
	// fmt.Println(conf)

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
		fmt.Println(obj)
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

		fmt.Printf("The exchange for %s is %s\n", wallet.Ticker, coinData.Tag) // TODO Replace with exchange factory method
		// TODO Optimize sendind TXs for the same coin (instead of making 5 dash transactions, make one)
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
