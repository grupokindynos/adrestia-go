package exchanges

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/grupokindynos/adrestia-go/exchanges/config"

	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/exchange_models"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/adrestia-go/utils"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/common/hestia"
	"github.com/joho/godotenv"
	"moul.io/http2curl"
)

type Cryptobridge struct {
	Exchange
	AccountName    string
	BitSharesUrl   string
	MasterPassword string
	Obol obol.ObolService
}

var CBInstance = NewCryptobridge()

func NewCryptobridge() *Cryptobridge {
	c := new(Cryptobridge)
	c.Name = "Cryptobridge"
	data := c.getSettings()
	c.MasterPassword = data.MasterPassword
	c.AccountName = data.AccountName
	c.BaseUrl = data.BaseUrl
	c.BitSharesUrl = data.BitSharesUrl
	return c
}

func (c *Cryptobridge) GetName() (string, error) {
	return c.Name, nil
}

func (c *Cryptobridge) GetAddress(coin coins.Coin) (string, error) {
	client := &http.Client{}
	url := "v2/accounts/" + c.AccountName + "/assets/" + coin.Tag + "/addresses"
	req, _ := http.NewRequest("GET", c.BitSharesUrl+url, nil)
	// ...
	req.Header.Add("key", "asjldfajsdlkfjasldflasjdfl")
	res, _ := client.Do(req)

	fmt.Println(res)
	return "Missing Implementation", nil
}

func (c *Cryptobridge) OneCoinToBtc(coin coins.Coin) (float64, error) {
	var rates = new(exchange_models.CBRates)
	url := "v1/ticker"
	err := getBitSharesRequest(c.MasterPassword, c.BaseUrl+url, http.MethodGet, nil, &rates)

	if err != nil {
		return 0.0, err
	}

	var pair = coin.Tag + "_BTC"
	var pair2 = "BTC_" + coin.Tag

	var r = 0.0

	for _, rate := range *rates {
		if rate.ID == pair {
			r, _ = strconv.ParseFloat(rate.Last, 64)
			return r, nil
		}
		if rate.ID == pair2 {
			r, _ = strconv.ParseFloat(rate.Last, 64)
			return 1 / r, nil
		}
	}
	// log.Fatalln("Not implemented")
	return 0.0, errors.New("no rates found")
}

func (c *Cryptobridge) GetBalances() ([]balance.Balance, error) {
	s := fmt.Sprintf("Retrieving Balances for %s", c.Name)
	log.Println(s)
	var balances []balance.Balance
	var CBResponse = new(exchange_models.CBBalance)
	url := "balance"
	err := getBitSharesRequest(c.MasterPassword, c.BitSharesUrl+url, http.MethodGet, nil, &CBResponse)
	if err != nil {
		return balances, err
	}

	for _, asset := range CBResponse.Data {
		if strings.Contains(asset.Symbol, "BRIDGE.") {
			asset.Symbol = asset.Symbol[7:]
		}
		rate, _ := c.Obol.GetCoin2CoinRates("BTC", asset.Symbol)

		var b = balance.Balance{
			Ticker:             asset.Symbol,
			ConfirmedBalance:   asset.Amount,
			UnconfirmedBalance: 0.0,
			RateBTC:            rate,
			DiffBTC:            0,
			IsBalanced:         false,
		}
		balances = append(balances, b)
	}
	s = utils.GetBalanceLog(balances, c.Name)
	log.Println(s)
	return balances, nil
}

func (c *Cryptobridge) SellAtMarketPrice(sellOrder transaction.ExchangeSell) (bool, string, error) {
	s := fmt.Sprintf("Selling %f %s for %s in %s", sellOrder.Amount, sellOrder.FromCoin.Tag, sellOrder.ToCoin.Tag, c.Name)
	log.Println(s)
	// sellorders/BRIDGE.{sell.To.tag}/BRIDGE.{sell.From.tag}
	url := "sellorders/BRIDGE." + strings.ToUpper(sellOrder.ToCoin.Tag) + "/BRIDGE." + strings.ToUpper(sellOrder.FromCoin.Tag)
	var openOrders = new(exchange_models.Orders)

	err := getBitSharesRequest(c.MasterPassword, c.BitSharesUrl+url, http.MethodGet, nil, &openOrders)

	if err != nil {
		return false, "", err
	}

	calculatedPrice := 0.0
	auxAmount := 0.0
	copySellOrder := sellOrder.Amount
	amountToSell := 0.0
	index := 0

	// log.Println("Selling Order ", sellOrder.Amount, " ", sellOrder.FromCoin.Tag)

	for auxAmount < sellOrder.Amount {
		currentAsk := openOrders.Data.Asks[index]
		auxAmount += currentAsk.Base.Amount
		calculatedPrice = 1 / currentAsk.Price * 0.9999 // TODO I (Helios) don't quite understand this way of calculating.
		if currentAsk.Base.Amount < copySellOrder {
			amountToSell = currentAsk.Base.Amount
		} else {
			amountToSell = copySellOrder
		}
		copySellOrder -= currentAsk.Base.Amount
		index++
	}
	fmt.Print("Calculated Price: ", calculatedPrice)
	fmt.Println("Amount to Sell: ", amountToSell)

	// TODO Create selling order

	return true, "order id", nil
}

//Withdraw allows Adrestia to send money from exchanges to a valid address
func (c *Cryptobridge) Withdraw(coin coins.Coin, address string, amount float64) (bool, error) {
	var withdrawObj = exchange_models.CBWithdraw{
		Amount:  amount,
		Address: address,
	}

	data, err := json.Marshal(withdrawObj)
	fmt.Println("Data: ", data)
	if err != nil {
		panic("Couldn't serialize Order object.")
	}

	url := c.BitSharesUrl + "withdraw/" + strings.ToLower(coin.Tag)

	//client := &http.Client{}
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	req.Header.Add("key", c.MasterPassword)

	command, _ := http2curl.GetCurlCommand(req)
	fmt.Println(command)
	// res, _ := client.Do(req)

	// TODO Fix CB KYC to enable withdrawals

	// println("Res: ", res)
	return true, nil
}

func (c *Cryptobridge) GetRateByAmount(sell transaction.ExchangeSell) (float64, error) {
	return 0.0, errors.New("func not implemented")
}

func (c *Cryptobridge) GetOrderStatus(orderId string) (hestia.AdrestiaStatus, error) {
	return -1, errors.New("func not implemented")
}

func (c *Cryptobridge) getSettings() config.CBAuth {
	if err := godotenv.Load(); err != nil {
		log.Println(err)
	}
	var data config.CBAuth
	data.AccountName = os.Getenv("CB_ACCOUNT_NAME")
	data.BaseUrl = os.Getenv("CB_BASE_URL")
	data.MasterPassword = os.Getenv("CB_MASTER_PASSWORD")
	data.BitSharesUrl = os.Getenv("CB_BITSHARES_URL")
	return data
}

// Builds requests with the appropriate header and returns the content in the desired struct
func getBitSharesRequest(key string, url string, method string, body io.Reader, outType interface{}) error {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("key", key)
	res, _ := client.Do(req)

	bodyResp, e := ioutil.ReadAll(res.Body)
	if e != nil {
		err = e
		return e
	}
	e = json.Unmarshal(bodyResp, &outType)
	if e != nil {
		//log.Fatalln("Error in unmarshall: ",jsonErr)
		err = e
		return e
	}
	return nil
}
