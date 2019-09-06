package balance

import "fmt"

type HotWalletBalances struct {
	Status int `json:"status"`
	Data   []HotWalletBalance `json:"data"`
	Error interface{} `json:"error"`
}

type HotWalletBalance struct {
	Ticker  string  `json:"ticker"`
	Balance float64 `json:"balance"`
}

func (b HotWalletBalances) PrintBalances(){
	for i, _ := range b.Data {
		fmt.Sprintf("$f %s", b.Data[i].Balance, b.Data[i].Ticker)
	}
}