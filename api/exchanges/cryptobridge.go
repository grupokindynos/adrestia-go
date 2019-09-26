package exchanges

import (
	"encoding/json"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/bitshares"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/common/coin-factory/coins"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	var rates = new(bitshares.CBRates)
	url := "v1/ticker"
	getBitSharesRequest(c.BaseUrl + url, http.MethodGet, nil, &rates)

	var pair = coin.Tag + "_BTC"
	var pair2 = "BTC_" + coin.Tag

	var r = 0.0

	for _, rate := range *rates {
		if rate.ID == pair {
			r, _ = strconv.ParseFloat(rate.Last, 64)
			return r
		}
		if rate.ID == pair2 {
			r, _ = strconv.ParseFloat(rate.Last, 64)
			return 1/r
		}
	}
	log.Fatalln("Not implemented")
	return 0.0
}

func (c Cryptobridge) GetBalances(coin coins.Coin) []balance.Balance {
	var balances []balance.Balance
	var CBResponse = new(bitshares.CBBalance)
	url := "balance"
	getBitSharesRequest(c.BitSharesUrl + url, http.MethodGet, nil, &CBResponse)

	for _,asset := range CBResponse.Data {
		if strings.Contains(asset.Symbol, "BRIDGE.") {
			asset.Symbol = asset.Symbol[7:]
		}
		var balance = balance.Balance{
			Ticker:     asset.Symbol,
			Balance:    asset.Amount,
			RateBTC:    0,
			DiffBTC:    0,
			IsBalanced: false,
		}
		balances = append(balances, balance)
	}
	return balances
}

func (c Cryptobridge) SellAtMarketPrice(SellOrder transaction.ExchangeSell) bool {
	// sellorders/BRIDGE.{sell.To.tag}/BRIDGE.{sell.From.tag}
	url := "sellorders/BRIDGE." + strings.ToUpper(SellOrder.ToCoin.Tag) + "/BRIDGE." + strings.ToUpper(SellOrder.FromCoin.Tag)
	fmt.Println(c.BitSharesUrl + url)
	var openOrders = new(bitshares.Orders)
	getBitSharesRequest(c.BitSharesUrl + url, http.MethodGet, nil, &openOrders)

	fmt.Println(openOrders)
	panic("Not implemented")
}


// Builds requests with the appropriate header and returns the content in the desired struct
func getBitSharesRequest(url string, method string, body io.Reader, outType interface{}) interface{} {
	// fmt.Println(url)
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, body)
	req.Header.Add("key", "asjldfajsdlkfjasldflasjdfl")
	res, _ := client.Do(req)
	bodyResp, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatalln("Error reading body: ", readErr)
	}
	jsonErr := json.Unmarshal(bodyResp, &outType)
	if jsonErr != nil {
		fmt.Println(res)
		log.Fatalln("Error in unmarshall: ",jsonErr)
	}
	return outType
}