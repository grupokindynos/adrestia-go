package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"

	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/services"
)

const baseUrl string = "https://delphi.polispay.com/api/"
const confPath string = "conf"

var ratePvdr = services.RateProvider{}
var printDebugInfo = true

func main() {
	fmt.Println("Program Started")

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
	fmt.Println(conf)
	SortBalances(balances, conf)

}

func GetWalletBalances() []balance.Balance {
	fmt.Println("\tRetrieving Wallet Balances...")

	var balances balance.HotWalletBalances
	response, err := http.Get(baseUrl + "v2/wallets/balances")
	if err != nil {
		fmt.Println(err)
	}

	defer response.Body.Close()

	if err != nil {
		fmt.Println(err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		err := json.Unmarshal(data, &balances)
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("Finished Retrieving Balances")

	var updatedBalances []balance.Balance
	fmt.Println("\tRetrieving Wallet Rates...")
	for _, coin := range balances.Data {
		fmt.Print(coin)
		var newBalance = coin
		newBalance.RateBTC = ratePvdr.GetRate(newBalance.Ticker)
		updatedBalances = append(updatedBalances, newBalance)
	}
	fmt.Println(updatedBalances)
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
	fmt.Println(conf)

	var firebaseConfBalances = conf.ToMap()

	return firebaseConfBalances

}

func SortBalances(inputBalances []balance.Balance, conf map[string]balance.Balance) ([]balance.Balance, []balance.Balance) {
	// Sorts Balances

	var balancedWallets []balance.Balance
	var unbalancedWallets []balance.Balance

	for _, obj := range inputBalances {
		fmt.Println("Debugg", conf[obj.Ticker].Balance)
		obj.GetDiff(conf[obj.Ticker].Balance)
		if obj.IsBalanced {
			balancedWallets = append(balancedWallets, obj)
		} else {
			unbalancedWallets = append(unbalancedWallets, obj)
		}
	}

	sort.Sort(balance.ByDiff(balancedWallets))
	sort.Sort(balance.ByDiff(unbalancedWallets))

	fmt.Println("Info Sorting")
	fmt.Println("Unbalanced", unbalancedWallets)
	fmt.Println("Balanced", balancedWallets)

	return balancedWallets, unbalancedWallets
}
