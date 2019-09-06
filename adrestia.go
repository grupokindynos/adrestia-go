package main

import(
	"encoding/json"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"io/ioutil"
	"net/http"
)

var baseUrl string = "https://delphi.polispay.com/api/"
var printDebugInfo = true

func main() {
	fmt.Println("Program Started")

	// Gets balance from Hot Wallets
	var balances = GetWalletBalances()


	if printDebugInfo{
		fmt.Println("\t\tAvailable Coins")
		for i,_ := range balances.Data{
			fmt.Println("\t\t",balances.Data[i].Balance, balances.Data[i].Ticker)
		}
	}
}

func GetWalletBalances() balance.HotWalletBalances {
	fmt.Println("\tRetrieving Wallet Balances...")

	var balances balance.HotWalletBalances
	response, error := http.Get(baseUrl + "v2/wallets/balances")

	defer response.Body.Close()

	if error != nil {
		fmt.Println(error)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		err := json.Unmarshal(data, &balances)
		if err != nil{
			fmt.Println(error)
		}
	}
	return balances
}

func GetFBConfiguration() balance.HotWalletBalances {
	// Retrieves minimum set balance configuration from Firebase conf

	var fireBaseConfBalances = new(balance.HotWalletBalances)

	return *fireBaseConfBalances

}

func SortBalances() ([]balance.HotWalletBalance, []balance.HotWalletBalance ){
	// Sorts Balances

	var balancedWallets []balance.HotWalletBalance
	var unbalancedWallets []balance.HotWalletBalance

	return balancedWallets, unbalancedWallets
}