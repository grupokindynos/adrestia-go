package exchanges

import (
	"encoding/json"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/bitshares"
	"github.com/grupokindynos/common/coin-factory/coins"
	"io/ioutil"
	"log"
	"net/http"
)

type Cryptobridge struct {
	Exchange
	AccountName string
	BitSharesUrl string
}

func NewCryptobridge() *Cryptobridge {
	c := new(Cryptobridge)
	c.Name = "Cryptobridge"
	c.BaseUrl = "https://api.crypto-bridge.org/"
	c.AccountName = "lakshmi-87"
	c.BitSharesUrl = "http://178.128.179.29:5000/"
	return c
}

func (c Cryptobridge) GetAddress(coin coins.Coin) string {
	client := &http.Client{}
	url := "v2/accounts/" + c.AccountName + "/assets/" + coin.Tag + "/addresses"
	req, _ := http.NewRequest("GET", c.BitSharesUrl + url, nil)
	// ...
	req.Header.Add("key", "asjldfajsdlkfjasldflasjdfl")
	fmt.Println(c.BitSharesUrl + url)
	res, _ := client.Do(req)


	fmt.Println(res)
	return "Missing Implementation"
}

func (c Cryptobridge) OneCoinToBtc(coin coins.Coin) float64 {
	panic("Missing Implementation")
}

func (c Cryptobridge) GetBalance(coin coins.Coin) []balance.Balance {
	client := &http.Client{}
	url := "balance"
	req, _ := http.NewRequest("GET", c.BitSharesUrl + url, nil)
	// ...
	req.Header.Add("key", "asjldfajsdlkfjasldflasjdfl")
	fmt.Println(c.BitSharesUrl + url)
	res, _ := client.Do(req)
	fmt.Println(res)
	panic("Missing Implementation")
}

func (c Cryptobridge) GetBalances(coin coins.Coin) []balance.Balance {
	var balances []balance.Balance
	var CBResponse = new(bitshares.CBBalance)
	url := "balance"
	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodGet, c.BitSharesUrl + url, nil)
	req.Header.Add("key", "asjldfajsdlkfjasldflasjdfl")
	res, _ := client.Do(req)
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatalln(readErr)
	}
	jsonErr := json.Unmarshal(body, &CBResponse)
	if jsonErr != nil {
		log.Fatalln(jsonErr)
	}
	fmt.Println("Balances")
	for _,asset := range CBResponse.Data {
		var balance = balance.Balance{
			Ticker:     asset.Asset.Symbol,
			Balance:    asset.Amount,
			RateBTC:    0,
			DiffBTC:    0,
			IsBalanced: false,
		}

		fmt.Println(balance)
		balances = append(balances, balance)
	}
	return balances
}