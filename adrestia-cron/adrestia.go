package main

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/color"
	services2 "github.com/grupokindynos/adrestia-go/api/services"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/plutus"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/common/obol"

	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/services"
)

const fiatThreshold = 10.00 // USD

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	// TODO Disable and Enable Shift at star nd ending of the process
	color.Info.Tips("Program Started")

	// Gets balance from Hot Wallets
	var balances = GetWalletBalances()
	// Firebase Wallet Configuration
	// var conf = GetFBConfiguration("test_data/testConf.json", false)
	var confHestia, err = services.GetCoinConfiguration()

	if err != nil {
		log.Fatalln(err)
	}

	availableWallets, errors := NormalizeWallets(balances, confHestia)

	fmt.Println(availableWallets)
	fmt.Println(errors)

	var sendToExchanges []transaction.PTx
	// Evaluate wallets with exceeding amount
	for i, w := range availableWallets {
		fmt.Println("Retrieving for ", i, " ", w)
		if w.FirebaseConf.Balances.HotWallet < w.HotWalletBalance.Balance {
			tx := new(transaction.PTx)
			tx.FromCoin = w.HotWalletBalance.Ticker
			tx.ToCoin = w.HotWalletBalance.Ticker
			tx.Amount = w.FirebaseConf.Balances.HotWallet - w.HotWalletBalance.Balance
			tx.Rate = 1.0

			sendToExchanges = append(sendToExchanges, *tx)
		}
	}
	ef := new(services2.ExchangeFactory)
	// Send remaining amount to exchanges using plutus
	for _, tx := range sendToExchanges{
		fmt.Println("------------ TX-----------")
		fmt.Println(tx)
		coinInfo, err := coinfactory.GetCoin(tx.FromCoin)
		if err != nil {
			color.Error.Tips("%s", err)
			continue
		}
		ex, err := ef.GetExchangeByName(coinInfo.Rates.Exchange)
		if err != nil {
			color.Error.Tips("%s", err)
			continue
		}
		add, err :=ex.GetAddress(*coinInfo)
		if err != nil {
			color.Error.Tips("%s", err)
			continue
		}
		color.Info.Tips("Sending %.8f %s to its exchange at %s", tx.Amount, tx.FromCoin, add)
	}
	// fmt.Println(conf)
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

	}
	*/
}

func GetWalletBalances() []balance.Balance {
	flagAllRates := false
	log.Println("Retrieving Wallet Balances...")
	var rawBalances []balance.Balance
	availableCoins := coinfactory.Coins
	for _, coin := range availableCoins {
		res, err := plutus.GetWalletBalance(os.Getenv("PLUTUS_URL"), strings.ToLower(coin.Tag), os.Getenv("ADRESTIA_PRIV_KEY"), "adrestia", os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("PLUTUS_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
		if err != nil {
			fmt.Println("Plutus Service Error: ", err)
		}
		var t = res.Confirmed + res.Unconfirmed
		if t == 0.0 {
			fmt.Println(fmt.Sprintf("%.8f %s of a total of %.8f\t%.2f%%", res.Confirmed, coin.Tag, res.Confirmed + res.Unconfirmed, 0.0))
		} else {
			fmt.Println(fmt.Sprintf("%.8f %s of a total of %.8f\t%.2f%%", res.Confirmed, coin.Tag, res.Confirmed + res.Unconfirmed, 100*res.Confirmed/(res.Confirmed + res.Unconfirmed)))
		}

		// Create Balance Object
		b := balance.Balance{}
		b.Balance = res.Confirmed
		b.Ticker = coin.Tag
		rawBalances = append(rawBalances, b)
	}
	log.Println("Finished Retrieving Balances")

	var errRates []string

	var updatedBalances []balance.Balance
	log.Println("Retrieving Wallet Rates...")
	for _, coin := range rawBalances {
		var currentBalance = coin
		rate, err := obol.GetCoin2CoinRates("https://obol-rates.herokuapp.com/", "btc", currentBalance.Ticker)
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
// TODO Connect this with Hestia instead of Firebase and Test Data
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
	superavit := 0.0 // Exceeding amount in balanced wallets
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
	exchangeSet := make(map[string]bool)
	ef := services2.ExchangeFactory{}

	for i, tx := range pendingTransactions {
		color.Info.Tips(fmt.Sprintf("Performing tx %d: From %s to %s amounting for %.8f %s (%.8f BTC)", i+1, tx.FromCoin, tx.ToCoin, tx.Amount / tx.Rate, tx.FromCoin, tx.Amount))
		coin, err := coinfactory.GetCoin(tx.ToCoin)
		if err != nil{
			log.Println(err)
		}
		ex, err := ef.GetExchangeByCoin(*coin)
		exName, err := ex.GetName()
		if err != nil{
			log.Println(err)
		}
		_, ok := exchangeSet[exName]
		if !ok {
			exchangeSet[exName] = true
		}
	}

	// Optimization for txes to exchanges
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
			availableCoins[elem.Ticker] = balance.WalletInfoWrapper{
				HotWalletBalance: mapBalances[elem.Ticker],
				FirebaseConf:     mapConf[elem.Ticker],
			}
		}
	}
	return availableCoins, missingCoins
}