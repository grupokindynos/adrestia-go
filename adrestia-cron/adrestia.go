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
			color.Info.Tips(fmt.Sprintf("Wallet has %.8f %s \t(%.8f BTC)", balances[i].Balance, balances[i].Ticker, balances[i].BalanceBTC))
		}
	}
	// Firebase Wallet Configuration
	var conf = GetFBConfiguration("test_data/testConf.json", true)
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
	flagAllRates := false
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

	var errRates []string

	var updatedBalances []balance.Balance
	log.Println("\tRetrieving Wallet Rates...")
	for _, coin := range rawBalances {
		// fmt.Println(coin)
		var currentBalance = coin
		rate, err := obol.GetCoin2CoinRates("btc", currentBalance.Ticker)
		if err != nil{
			flagAllRates = true
			errRates = append(errRates, coin.Ticker)
		} else {
			currentBalance.SetRate(rate)
			updatedBalances = append(updatedBalances, currentBalance)
		}

	}
	if flagAllRates {
		color.Error.Tips("Not all rates could be retrieved. Balancing the rest of them. Missing rates for %s", errRates)
	}
	return updatedBalances
}

// Retrieves minimum set balance configuration from Firebase conf
func GetFBConfiguration(file string, load bool) map[string]balance.Balance {
	if load {
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
	} else {
		fireBase := services.InitFirebase()
		var conf balance.MinBalanceConfResponse
		conf, err := fireBase.GetConf()
		if err != nil {
			log.Fatal("Configuration not found")
		}
		var firebaseConfBalances = conf.ToMap()
		return firebaseConfBalances
	}
}

// Sorts Balances given their diff, so that topped wallets are used to fill the missing ones
func SortBalances(inputBalances []balance.Balance, conf map[string]balance.Balance) ([]balance.Balance, []balance.Balance) {
	var balancedWallets []balance.Balance
	var unbalancedWallets []balance.Balance

	for _, obj := range inputBalances {
		log.Println(fmt.Sprintf("SortBalances:: %.8f", conf[obj.Ticker].Balance))
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
	totalBtc := 0.0 // total amount in wallets

	for _, wallet := range balanced {
		superavit += math.Abs(wallet.DiffBTC)
		totalBtc += wallet.Balance * wallet.RateBTC
	}
	for _, wallet := range unbalanced {
		deficit += wallet.DiffBTC
		totalBtc += wallet.Balance * wallet.RateBTC
	}
	fmt.Println("Superavit: ", superavit)
	fmt.Println("Deficit: ", deficit)
	color.Info.Tips(fmt.Sprintf("Total in Wallets: %.8f", totalBtc))
	return superavit > math.Abs(deficit), superavit - math.Abs(deficit)
}

// Actual balancing action
func BalanceHW(balanced []balance.Balance, unbalanced []balance.Balance) []transaction.PTx {
	var pendingTransactions []transaction.PTx
	i := 0 // Balanced wallet index
	for _, wallet := range unbalanced {
		log.Println("BalanceHW:: Balancing ", wallet.Ticker)
		filledAmount := 0.0 // Amount that stores current fulfillment of a Balancing Transaction.
		initialDiff := math.Abs(wallet.DiffBTC)
		// TODO add fields for out addresses
		// coinData, _ := CoinFactory.GetCoin(wallet.Ticker)
		for filledAmount < initialDiff {
			color.Info.Tips(fmt.Sprintf("BalanceHW::\tUsing %s to balanace %s", balanced[i].Ticker, wallet.Ticker))
			if balanced[i].DiffBTC < initialDiff - filledAmount {
				var newTx =  transaction.PTx{
					ToCoin:  wallet.Ticker,
					FromCoin: balanced[i].Ticker,
					Amount: balanced[i].DiffBTC,
					Rate: balanced[i].RateBTC,
				}
				filledAmount += balanced[i].DiffBTC
				balanced[i].DiffBTC = 0.0
				i++
				fmt.Println(newTx)
				pendingTransactions = append(pendingTransactions, newTx)
			}else {
				filledAmount += initialDiff - filledAmount
				balanced[i].DiffBTC -= initialDiff - filledAmount
				var newTx =  transaction.PTx{
					ToCoin: wallet.Ticker,
					FromCoin: balanced[i].Ticker,
					Amount: initialDiff - filledAmount,
					Rate: balanced[i].RateBTC,
				}

				pendingTransactions = append(pendingTransactions, newTx)
			}
			// TODO Optimize sending TXs for the same coin (instead of making 5 dash transactions, make one)
			//color.Info.Tips(fmt.Sprintf("The exchange for %s is %s", wallet.Ticker, coinData.Rates.Exchange))
		}
	}
	fmt.Println(pendingTransactions)
	for i, tx := range pendingTransactions {
		color.Info.Tips(fmt.Sprintf("Performing tx %d: From %s to %s amounting for %.8f %s (%.8f BTC)", i+1, tx.FromCoin, tx.ToCoin, tx.Amount / tx.Rate, tx.FromCoin, tx.Amount))
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
