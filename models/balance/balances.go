package balance

import (
	"fmt"
)

type HotWalletBalances struct {
	Status int        `json:"status"`
	Data   []Balance  `json:"data"`
	Error interface{} `json:"error"`
}

type Balance struct {
	Ticker  string  	`json:"ticker"`
	Balance float64 	`json:"balance"`
	RateBTC float64		`json:"rateBTC"`
	DiffBTC float64		`json:"diffBTC"`
	IsBalanced bool		`json:"isBalanced"`
}

func (b Balance) GetDiff(target float64){
	b.DiffBTC = (target - b.Balance) * b.RateBTC
	fmt.Print("Njadnsaidbsa ", b.Ticker, " ", b.DiffBTC)
	if b.DiffBTC >= 0.0 { b.IsBalanced = true }else { b.IsBalanced = false }

}

// Sort Struct
type ByDiff []Balance

func (a ByDiff) Len() int           { return len(a) }
func (a ByDiff) Less(i, j int) bool { return a[i].DiffBTC < a[j].DiffBTC }
func (a ByDiff) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (b HotWalletBalances) PrintBalances(){
	for i, _ := range b.Data {
		fmt.Sprintf("$f %s", b.Data[i].Balance, b.Data[i].Ticker)
	}
}

// Coins Must be addded here
// TODO Update docs to indicate new coins must be added here
type MinBalanceConfResponse struct {
	BTC Balance `json:"BTC"`
	COLX Balance `json:"COLX"`
	DASH Balance `json:"DASH"`
	DGB Balance `json:"DGB"`
	GRS Balance `json:"GRS"`
	LTC Balance `json:"LTC"`
	POLIS Balance `json:"POLIS"`
	XZC Balance `json:"XZC"`
}

// Gets map for Ticker to Balance Object
// TODO Automate this part by using reflect
func (br MinBalanceConfResponse) ToMap() map[string]Balance {
	// Map structure
	var balanceMap = make(map[string]Balance)
	balanceMap["BTC"] = br.BTC
	balanceMap["COLX"] = br.COLX
	balanceMap["DASH"] = br.DASH
	balanceMap["DGB"] = br.DGB
	balanceMap["GRS"] = br.GRS
	balanceMap["LTC"] = br.LTC
	balanceMap["POLIS"] = br.POLIS
	balanceMap["XZC"] = br.XZC
	fmt.Println(balanceMap)
	return balanceMap
}

type MinBalanceConf struct {

}