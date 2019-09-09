package main

import (
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/services"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"net/http"
)

const baseUrl string = "https://delphi.polispay.com/api/"
const confPath string = "conf"
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

	var conf = GetFBConfiguration()
	fmt.Println(conf)
	SortBalances(balances, conf)

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
	fmt.Println("Finished Retrieving Balances")
	return balances
}

// Retrieves minimum set balance configuration from Firebase conf
func GetFBConfiguration() map[string]balance.Balance {
	// service account credentials
	opt := option.WithCredentialsFile("./fb_conf.json")
	config := &firebase.Config{
		DatabaseURL: "https://polispay-copay.firebaseio.com",
	}
	firebaseApp, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		panic(err)
	}
	fireBase := services.InitFirebase(firebaseApp)
	var conf balance.MinBalanceConfResponse
	conf, err = fireBase.GetConf()
	if err != nil {
		log.Fatal("Configuration not found")
	}
	fmt.Println(conf)

	var firebaseConfBalances = conf.ToMap()

	return firebaseConfBalances

}

func SortBalances(inputBalances balance.HotWalletBalances, conf map[string]balance.Balance) ([]balance.Balance, []balance.Balance){
	// Sorts Balances

	var balancedWallets []balance.Balance
	var unbalancedWallets []balance.Balance

	for i, obj := range inputBalances.Data{
		fmt.Println(i, obj)
	}

	return balancedWallets, unbalancedWallets
}
