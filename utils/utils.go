package utils

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"sort"

	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/common/hestia"
	cObol "github.com/grupokindynos/common/obol"
)

var obol cObol.ObolRequest
var minimumUSDTxAmount = 70.0

// This function normalizes the wallets that were detected in Plutus and those with configuration in Hestia.
// Returns a map of the coins' ticker as key containing a wrapper with both the actual balance of the wallet and
// its firebase configuration.
func NormalizeWallets(balances []balance.Balance, hestiaConf []hestia.Coin) (map[string]balance.WalletInfoWrapper, []string) {
	var activeCoins = make(map[string]bool) // TODO Replace with Hestia call
	for _, coin := range hestiaConf {
		if coin.Adrestia {
			activeCoins[coin.Ticker] = true
		}
	}

	fmt.Printf("balances %+v\n", balances)
	fmt.Printf("hestiaConf %+v\n", hestiaConf)

	var mapBalances = make(map[string]balance.Balance)
	var mapConf = make(map[string]hestia.Coin)
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
				if _, ok := activeCoins[elem.Ticker]; ok {
					availableCoins[elem.Ticker] = balance.WalletInfoWrapper{
						HotWalletBalance: elem,
						FirebaseConf:     mapConf[elem.Ticker],
					}
				}
			}
		}
	}
	fmt.Println("Balancing these Wallets: ", availableCoins)
	fmt.Println("Errors on these Wallets: ", missingCoins)
	return availableCoins, missingCoins
}

func SortBalances(data map[string]balance.WalletInfoWrapper) (balancedWallets []balance.Balance, unbalancedWallets []balance.Balance) {
	/*
		Sorts Balances given their diff, so that topped wallets are used to fill the missing ones
	*/

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

	log.Printf("Sorted Wallets")
	for _, wallet := range unbalancedWallets {
		fmt.Printf("%s has a deficit of %.8f BTC\n", wallet.Ticker, wallet.DiffBTC)
	}
	for _, wallet := range balancedWallets {
		fmt.Printf("%s has a superavit of %.8f BTC\n", wallet.Ticker, wallet.DiffBTC)
	}
	return balancedWallets, unbalancedWallets
}

func BalanceHW(balanced []balance.Balance, unbalanced []balance.Balance) ([]transaction.PTx, error) {
	var pendingTransactions []transaction.PTx
	var newTx transaction.PTx
	var amountTx float64
	var rateBTC float64
	rateUSD, err := getUSDRate("BTC")
	if err != nil {
		return pendingTransactions, err
	}

	i := 0 // Balanced wallet index
	for _, wallet := range unbalanced {
		log.Println("BalanceHW:: Balancing ", wallet.Ticker)
		diff := math.Abs(wallet.DiffBTC)
		amountUSD := diff * rateUSD
		for amountUSD >= minimumUSDTxAmount {
			if balanced[i].DiffBTC*rateUSD < minimumUSDTxAmount {
				i++
				continue
			}
			if balanced[i].DiffBTC >= diff {
				amountTx = diff
				rateBTC = balanced[i].RateBTC
				balanced[i].DiffBTC -= diff
			} else {
				amountTx = balanced[i].DiffBTC
				rateBTC = balanced[i].RateBTC
				balanced[i].DiffBTC = 0.0
			}
			newTx.Amount = amountTx / rateBTC
			newTx.ToCoin = wallet.Ticker
			newTx.FromCoin = balanced[i].Ticker
			newTx.BtcRate = rateBTC
			pendingTransactions = append(pendingTransactions, newTx)
			diff -= amountTx
			amountUSD = diff * rateUSD
			if balanced[i].DiffBTC == 0.0 {
				i++
			}
		}
	}
	return pendingTransactions, nil
}

func getUSDRate(coin string) (float64, error) {
	obol.ObolURL = os.Getenv("OBOL_URL")
	obolRates, err := obol.GetCoinRates(coin)
	if err != nil {
		return 0.0, err
	}

	for _, rate := range obolRates {
		if rate.Code == "USD" {
			return rate.Rate, nil
		}
	}

	return 0.0, errors.New("USD rate not found")
}

func DetermineBalanceability(balanced []balance.Balance, unbalanced []balance.Balance) (bool, float64) {
	superavit := 0.0 // Exceeding amount in balanced wallets
	deficit := 0.0   // Missing amount in unbalanced wallets
	totalBtc := 0.0  // total amount in wallets

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
	return superavit >= math.Abs(deficit), superavit - math.Abs(deficit)
}
