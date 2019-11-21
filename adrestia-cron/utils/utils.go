package utils

import (
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/common/hestia"
	"sort"
)

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
			// fmt.Println(elem.Ticker, "\n", mapConf[elem.Ticker].Balances.HotWallet)
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

func SortBalances(data map[string]balance.WalletInfoWrapper) ([]balance.Balance, []balance.Balance) {
	/*
		Sorts Balances given their diff, so that topped wallets are used to fill the missing ones
	*/
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