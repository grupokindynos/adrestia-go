package balance

import (
	"fmt"
	"github.com/grupokindynos/common/hestia"
	"log"
)

type HotWalletBalances struct {
	Status int         `json:"status"`
	Data   []Balance   `json:"data"`
	Error  interface{} `json:"error"`
}

type Balance struct {
	Ticker     				string  	`json:"ticker"`
	ConfirmedBalance    	float64 	`json:"confirmedBalance"`
	UnconfirmedBalance  	float64 	`json:"unconfirmedBalance"`
	Target 					float64		`json:"target"`
	RateBTC    				float64 	`json:"rateBTC"`
	DiffBTC   				float64 	`json:"diffBTC"`
	IsBalanced 				bool    	`json:"isBalanced"`
}

func (b *Balance) GetConfirmedProportion() float64 {
	if b.ConfirmedBalance + b.UnconfirmedBalance == 0 {
		return 0.0
	}
	return b.ConfirmedBalance * 100.0 / (b.ConfirmedBalance + b.UnconfirmedBalance)
}

func (b *Balance) GetBalanceInBtc(totalBalance bool) float64 {
	/*
		Returns the balance in BTC
		if totalBalance is set to TRUE returns the conversion using the amount the exchange is expecting so it is not recommended
		otherwise it returns only the CONFIRMED balance at the exchange.
	*/
	if b.RateBTC == 0.0 {
		return 0.0;
	}
	if totalBalance {
		return (b.ConfirmedBalance + b.UnconfirmedBalance) * b.RateBTC
	}
	return b.ConfirmedBalance * b.RateBTC
}

func (b *Balance) GetTotalBalance() float64{
	return b.UnconfirmedBalance + b.ConfirmedBalance
}

func (b *Balance) GetDiff() {
	log.Println(fmt.Sprintf("%s has %.8f as balance, a target of %.8f and a rate of %.8f", b.Ticker, b.GetTotalBalance(), b.Target, b.RateBTC))
	b.DiffBTC = (b.GetTotalBalance() - b.Target) * b.RateBTC
	// TODO Update this section to account for Tx/Miner Fees
	// Make it >= a range
	if b.DiffBTC >= 0.0 {
		b.IsBalanced = true
	} else {
		b.IsBalanced = false
	}
}

func (b *Balance) SetTarget(target float64) {
	b.Target = target
}

func (b *Balance) SetRate(rate float64){
	b.RateBTC = rate
}

// Sort Struct
type ByDiff []Balance

func (a ByDiff) Len() int           { return len(a) }
func (a ByDiff) Less(i, j int) bool {
	if a[i].Ticker == "BTC" {
		return true
	} else {
		return a[i].DiffBTC < a[j].DiffBTC
	}
}
func (a ByDiff) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByDiffInverse []Balance

func (a ByDiffInverse) Len() int           { return len(a) }
func (a ByDiffInverse) Less(i, j int) bool {
	if a[i].Ticker == "BTC" {
		return false
	} else {
		return a[i].DiffBTC > a[j].DiffBTC
	}

}
func (a ByDiffInverse) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (b HotWalletBalances) PrintBalances() {
	for i, _ := range b.Data {
		fmt.Printf("%f %s", b.Data[i].ConfirmedBalance, b.Data[i].Ticker)
	}
}

type WalletInfoWrapper struct {
	HotWalletBalance 	Balance
	FirebaseConf 		hestia.Coin
}
