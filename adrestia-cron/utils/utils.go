package utils

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/adrestia-go/services"
	"github.com/grupokindynos/common/hestia"
	"log"
	"math"
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

func ChangeOrderStatus(order hestia.AdrestiaOrder, status hestia.AdrestiaStatus) () {
	fallbackStatus := order.Status
	order.Status = hestia.AdrestiaStatusStr[status]
	resp, err := services.UpdateAdrestiaOrder(order)
	// TODO Move in map (if concurrency on maps allows for it)
	if err != nil {
		order.Status = fallbackStatus
		fmt.Println(err)
	} else {
		log.Println(fmt.Sprintf("order %s in %s has been updated to %s\t%s", order.OrderId, order.Exchange, order.Status, resp))
	}
}

func StoreOrders(orders []hestia.AdrestiaOrder) {
	for _, order := range orders {
		res, err := services.CreateAdrestiaOrder(order)
		if err != nil {
			fmt.Println("error posting order to hestia: ", err)
		} else {
			fmt.Println(res)
		}
	}
}

func BalanceHW(balanced []balance.Balance, unbalanced []balance.Balance) []transaction.PTx {
	/*
		DEPRECATED Used to balance adrestia. Keeping for reference.
	 */
	var pendingTransactions []transaction.PTx
	i := 0 // Balanced wallet index
	for _, wallet := range unbalanced {
		// log.Println("BalanceHW:: Balancing ", wallet.Ticker)
		filledAmount := 0.0 // Amount that stores current fulfillment of a Balancing Transaction.
		initialDiff := math.Abs(wallet.DiffBTC)
		// TODO add fields for out addresses
		for filledAmount < initialDiff {
			// color.Info.Tips(fmt.Sprintf("BalanceHW::\tUsing %s to balanace %s", balanced[i].Ticker, wallet.Ticker))
			var newTx transaction.PTx
			if balanced[i].DiffBTC < initialDiff - filledAmount {
				newTx.ToCoin = wallet.Ticker
				newTx.FromCoin = balanced[i].Ticker
				newTx.Amount = balanced[i].DiffBTC
				newTx.Rate = balanced[i].RateBTC

				filledAmount += balanced[i].DiffBTC
				balanced[i].DiffBTC = 0.0
				i++
				fmt.Println("Type I tx: ", newTx)
				pendingTransactions = append(pendingTransactions, newTx)
			}else {
				newTx.Amount = initialDiff - filledAmount
				filledAmount += initialDiff - filledAmount
				balanced[i].DiffBTC -= initialDiff - filledAmount
				newTx.ToCoin = wallet.Ticker
				newTx.FromCoin = balanced[i].Ticker
				newTx.Rate = balanced[i].RateBTC
				fmt.Println("Type II tx: ", newTx)
				pendingTransactions = append(pendingTransactions, newTx)
			}
		}
	}
	return pendingTransactions
}

func DetermineBalanceability(balanced []balance.Balance, unbalanced []balance.Balance) (bool, float64) {
	/*
		DEPRECATED Used to balance adrestia. Keeping for reference.
	 */
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