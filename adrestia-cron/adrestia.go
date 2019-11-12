package main

import (
	"encoding/json"
	"fmt"
	"github.com/gookit/color"
	apiServices "github.com/grupokindynos/adrestia-go/api/services"
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

const fiatThreshold = 2.00 // USD // 2.0 for Testing, 10 USD for production

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	balOrders, err := services.GetBalancingOrders()
	if err != nil {
		fmt.Print(err)
		panic(err)
	} else {
		fmt.Println("Balancing Orders: ", balOrders)
	}
	// TODO Disable and Enable Shift at star nd ending of the process
	color.Info.Tips("Program Started")
	var balances = GetWalletBalances()						// Gets balance from Hot Wallets
	confHestia, err := services.GetCoinConfiguration()	// Firebase Wallet Configuratio
	if err != nil {
		log.Fatalln(err)
	}
	availableWallets, error_wallets := NormalizeWallets(balances, confHestia) // Verifies wallets in firebase are the same as in plutus and creates a map

	fmt.Println("Balancing these Wallets: ", availableWallets)
	fmt.Println("Errors on these Wallets: ", error_wallets)

	// TODO Sort
	balanced, unbalanced := SortBalances(availableWallets)
	status, amount := DetermineBalanceability(balanced, unbalanced)
	log.Println(fmt.Sprintf("Wallets Balanceability Status: %t\nAmount (+/-): %.8f", status, amount))

	// Calculates Balancing txes
	var sendToExchanges []transaction.PTx
	// Evaluate wallets with exceeding amount
	for i, w := range availableWallets {
		fmt.Println("Retrieving for ", i, " ", w)
		if w.FirebaseConf.Balances.HotWallet < w.HotWalletBalance.GetTotalBalance() {
			tx := new(transaction.PTx)
			tx.FromCoin = w.HotWalletBalance.Ticker
			tx.ToCoin = w.HotWalletBalance.Ticker
			tx.Amount = w.FirebaseConf.Balances.HotWallet - w.HotWalletBalance.GetTotalBalance()
			tx.Rate = 1.0
			sendToExchanges = append(sendToExchanges, *tx)
		}
	}
	ef := new(apiServices.ExchangeFactory)
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

	}*/
}

func GetWalletBalances() []balance.Balance {
	flagAllRates := false
	log.Println("Retrieving Wallet Balances...")
	var rawBalances []balance.Balance
	availableCoins := coinfactory.Coins
	for _, coin := range availableCoins {
		res, err := plutus.GetWalletBalance(os.Getenv("PLUTUS_URL"), strings.ToLower(coin.Tag), os.Getenv("ADRESTIA_PRIV_KEY"), "adrestia", os.Getenv("PLUTUS_AUTH_USERNAME"), os.Getenv("PLUTUS_AUTH_PASSWORD"), os.Getenv("PLUTUS_PUBLIC_KEY"), os.Getenv("MASTER_PASSWORD"))
		if err != nil {
			fmt.Println(fmt.Sprintf("Plutus Service Error for %s: %v", coin.Tag, err))
		} else {
			// Create Balance Object
			b := balance.Balance{}
			b.ConfirmedBalance = res.Confirmed
			b.UnconfirmedBalance = res.Unconfirmed
			b.Ticker = coin.Tag
			rawBalances = append(rawBalances, b)
			fmt.Println(fmt.Sprintf("%.8f %s\t of a total of %.8f\t%.2f%%", b.ConfirmedBalance, b.Ticker, b.ConfirmedBalance + b.UnconfirmedBalance, b.GetConfirmedProportion()))
		}
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
		x:= obj.HotWalletBalance // Curreny Wallet Info Wrapper
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
	ef := apiServices.ExchangeFactory{}

	for i, tx := range pendingTransactions {
		color.Info.Tips(fmt.Sprintf("Performing tx %d: From %s to %s amounting for %.8f %s (%.8f BTC)", i+1, tx.FromCoin, tx.ToCoin, tx.Amount / tx.Rate, tx.FromCoin, tx.Amount))
		coin, err := coinfactory.GetCoin(tx.ToCoin)
		if err != nil{
			log.Println(err)
		}
		ex, err := ef.GetExchangeByCoin(*coin)
		if err == nil {
			exName, err := ex.GetName()
			if err != nil{
				log.Println(err)
			}
			_, ok := exchangeSet[exName]
			if !ok {
				exchangeSet[exName] = true
			}
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