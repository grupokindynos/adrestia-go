package exchanges

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/common/hestia"
)

//Bithumb The attributes needed for the Bithumb exchanges
type Bithumb struct {
	Name          string
	user          string
	authorization string
	url           string
	addresses     map[string]string
	client        *http.Client
}

//NewBithumb Creates a new instance of Bithumb
func NewBithumb(params models.ExchangeParams) *Bithumb {
	b := new(Bithumb)
	b.Name = params.Name
	b.user = params.Keys.PublicKey
	b.authorization = params.Keys.PrivateKey
	b.client = &http.Client{}
	b.addresses = map[string]string{
		"BTC": "1L9jPKCbUbK9aKgn5miwRmUy51Pm64SeW6",
		"GTH": "",
		"USDT": "",
	}
	b.url = "https://global-openapi.bithumb.pro/openapi/v1"
	return b
}

// GetName Gets the exchange name
func (b *Bithumb) GetName() (string, error) {
	return b.Name, nil
}

// GetAddress Gets address from Bithumb WIP: Add error message
func (b *Bithumb) GetAddress(asset string) (string, error) {
	return b.addresses[asset], nil
}

// GetBalance Gets the balance for a given asset
func (b *Bithumb) GetBalance(asset string) (float64, error) {
	assetInfo, err := b.Assets(asset)
	if err != nil {
		return 0, err
	}
	if len(assetInfo.Data) >= 1 {
		balance, _ := assetInfo.Data[0].Count.Float64()
		return balance, nil
	}
	return 0, errors.New("asset not found")
}

func (b *Bithumb) SellAtMarketPrice(order hestia.Trade) (string, error) {
	orderInfo, err := b.createOrder(order.Symbol, strings.ToLower(order.Side), decimal.NewFromFloat(order.Amount), decimal.NewFromFloat(0), strings.ToLower("market"))
	if err != nil {
		return "", err
	}
	return orderInfo.Data.OrderId, nil
}

func (b *Bithumb) Withdraw(asset string, address string, amount float64) (string, error) {
	_, err := b.withdraw(asset, address, decimal.NewFromFloat(amount), "")
	if err != nil {
		return "", err
	}
	return "", nil
}

func (b *Bithumb) GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error) {
	orderStatus, err := b.orderDetail(order.Symbol, order.OrderId)
	if err != nil {
		return hestia.ExchangeOrderInfo{
			Status:         hestia.ExchangeOrderStatusOpen,
			ReceivedAmount: 0,
		}, err
	}
	switch orderStatus.Data.Status {
	case "send":
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusOpen,
			ReceivedAmount: 0,
		}, nil
	case "pending":
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusOpen,
			ReceivedAmount: 0,
		}, nil
	case "success":
		tradedValue, _ := orderStatus.Data.TradedNum.Float64()
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusCompleted,
			ReceivedAmount: tradedValue,
		}, nil
	case "cancel":
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusError,
			ReceivedAmount: 0,
		}, nil
	default:
		return hestia.ExchangeOrderInfo{
			Status: hestia.ExchangeOrderStatusOpen,
			ReceivedAmount: 0,
		}, nil
	}
}

func (b *Bithumb) GetPair(fromCoin string, toCoin string) (models.TradeInfo, error) {
	markets, err := b.getConfig()
	if err != nil {
		return models.TradeInfo{}, err
	}
	symbolBuy := fmt.Sprintf("%s-%s", strings.ToUpper(fromCoin), strings.ToUpper(toCoin))
	symbolSell := fmt.Sprintf("%s-%s", strings.ToUpper(toCoin), strings.ToUpper(fromCoin))
	for _, spotData := range markets.Data.SpotConfig {
		if spotData.Symbol == symbolBuy {
			return models.TradeInfo{
				Book: spotData.Symbol,
				Type: "buy",
			}, nil
		} else if spotData.Symbol == symbolSell{
			return models.TradeInfo{
				Book: spotData.Symbol,
				Type: "sell",
			}, nil
		}
	}
	t := models.TradeInfo{}
	return t, errors.New("pair not found")
}

func (b *Bithumb) GetWithdrawalTxHash(txId string, asset string) (string, error) {
	// No way of knowing a withdrawal tx hash
	return "", nil
}

//  Gets the deposit status from an asset's exchange.
func (b *Bithumb) GetDepositStatus(addr string, txId string, asset string) (hestia.ExchangeOrderInfo, error) {
	// bithumb does not provide a way of searching for this, maybe we can use block explorer to determine XY network confirmations
	e := hestia.ExchangeOrderInfo{}
	return e, nil
}

// Functions to get Bithumb API data

type configResp struct {
	baseResp
	Data struct {
		CoinConfig []struct {
			MakerFeeRate   decimal.Decimal `json:"makerFeeRate"`
			MinWithdraw    decimal.Decimal `json:"minWithdraw"`
			WithdrawFee    decimal.Decimal `json:"withdrawFee"`
			Name           string `json:"name"`
			DepositStatus  string `json:"depositStatus"`
			FullName       string `json:"fullName"`
			TakerFeeRate   decimal.Decimal `json:"takerFeeRate"`
			WithdrawStatus decimal.Decimal `json:"withdrawStatus"`
		} `json:"coinConfig"`
		ContractConfig []struct {
			Symbol       string `json:"symbol"`
			MakerFeeRate decimal.Decimal `json:"makerFeeRate"`
			TakerFeeRate decimal.Decimal `json:"takerFeeRate"`
		} `json:"contractConfig"`
		SpotConfig []struct {
			Symbol       string   `json:"symbol"`
			Accuracy     []string `json:"accuracy"`
			PercentPrice struct {
				MultiplierDown decimal.Decimal `json:"multiplierDown"`
				MultiplierUp   decimal.Decimal `json:"multiplierUp"`
			} `json:"percentPrice"`
		} `json:"spotConfig"`
	} `json:"data"`
}

type createOrderResp struct {
	baseResp
	Data struct {
		OrderId string
		Symbol  string
	}
}

type orderDetailResp struct {
	baseResp
	Data struct {
		OrderID    string `json:"orderId"`
		Symbol     string `json:"symbol"`
		Price      decimal.Decimal `json:"price"`
		TradedNum  decimal.Decimal `json:"tradedNum"`
		Quantity   decimal.Decimal `json:"quantity"`
		AvgPrice   decimal.Decimal `json:"avgPrice"`
		Status     string `json:"status"`
		Type       string `json:"type"`
		Side       string `json:"side"`
		CreateTime string `json:"createTime"`
		TradeTotal decimal.Decimal `json:"tradeTotal"`
	} `json:"data"`
}

type assetsResp struct {
	baseResp
	Data []struct {
		CoinType    string
		Count       decimal.Decimal
		Frozen      decimal.Decimal
		Type        string
		BtcQuantity decimal.Decimal
	}
}

type baseResp struct {
	Code      string
	Msg       string
	Timestamp int64
	Data      interface{}
}

func handleErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func (b *Bithumb) getSha256HashCode(preSign string) string {
	h := hmac.New(sha256.New, []byte(b.authorization))
	h.Write([]byte(preSign))
	hashCode := hex.EncodeToString(h.Sum(nil))
	return hashCode
}

func (b *Bithumb) sign(preMap map[string]string) string {
	var keys []string
	for k := range preMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var preSign string
	for _, k := range keys {
		preSign += k + "=" + preMap[k] + "&"
	}
	preSign = strings.TrimSuffix(preSign, "&")
	fmt.Println("prepare signature string >======= ", preSign)
	signature := b.getSha256HashCode(preSign)
	fmt.Println("signature string >====== ", signature)
	return signature
}

func (b *Bithumb) post(url string, params interface{}, result interface{}) error {
	preMap := b.struct2map(params)
	preMap["apiKey"] = b.user
	preMap["timestamp"] = strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	preMap["signature"] = b.sign(preMap)
	err := post(url, preMap, result)
	handleErr(err)
	return err
}

func (b *Bithumb) struct2map(params interface{}) map[string]string {
	t := reflect.TypeOf(params)
	v := reflect.ValueOf(params)
	var data = make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Tag.Get("json")] = v.Field(i).String()
	}
	return data
}

func (b *Bithumb) get(url string, r interface{}) error {
	resp := doGet(url)
	err := doParse(resp, r)
	return err
}

func doGet(url string) []byte {
	resp, err := http.Get(url)
	return handleResp(resp, err)
}

func post(url string, params interface{}, r interface{}) error {
	jsonBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}
	resp := doPost(url, jsonBytes)
	nil := doParse(resp, r)
	return nil
}

func doParse(resp []byte, in interface{}) error {
	err := json.Unmarshal(resp, in)
	if err != nil {
		return err
	}
	return nil
}

func doPost(url string, data []byte) []byte {
	body := bytes.NewReader(data)
	resp, err := http.Post(url, "application/json", body)
	return handleResp(resp, err)
}

func handleResp(resp *http.Response, err error) []byte {
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return r
}

/** API Implementation **/
func (b *Bithumb) Assets(coinType string) (*assetsResp, error) {
	var r assetsResp
	p := struct {
		CoinType  string `json:"coinType"`
		AssetType string `json:"assetType"`
	}{
		coinType, "spot",
	}
	err := b.post(b.url+"/spot/assetList", p, &r)
	if err != nil {
		return &r, err
	}
	return &r, nil
}

func (b *Bithumb) withdraw(asset string, address string, quantity decimal.Decimal, mark string) (bool, error) {
	var r assetsResp
	p := struct {
		CoinType string `json:"coinType"`
		Address  string `json:"address"`
		Quantity string `json:"quantity"`
		Mark     string `json:"mark"`
	}{
		asset, address, quantity.String(), mark,
	}
	err := b.post(b.url+"/spot/assetList", p, &r)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (b *Bithumb) createOrder(symbol string, side string, quantity decimal.Decimal, price decimal.Decimal, orderType string) (*createOrderResp, error) {
	var c createOrderResp
	var pr = price
	if orderType == "market" {
		pr = decimal.NewFromFloat(-1)
	}
	p := struct {
		Symbol    string `json:"symbol"`
		Type      string `json:"type"`
		Side      string `json:"side"`
		Price     string `json:"price"`
		Quantity  string `json:"quantity"`
		Timestamp string `json:"timestamp"`
	}{
		symbol, orderType, side, pr.String(), quantity.String(), strconv.FormatInt(time.Now().UTC().UnixNano()/1e6, 10),
		// symbol, orderType, side, pr.String(), quantity.String(),
	}
	err := b.post(b.url+"/spot/placeOrder", p, &c)
	if err != nil {
		return &c, err
	}
	fmt.Println("message: ", c.Msg)
	return &c, nil
}

func (b *Bithumb) orderDetail(symbol string, orderId string) (*orderDetailResp, error) {
	var c orderDetailResp
	p := struct {
		Symbol  string `json:"symbol"`
		OrderId string `json:"orderId"`
	}{
		symbol, orderId,
	}
	err := b.post(b.url+"/spot/singleOrder", p, &c)
	if err != nil {
		return &c, err
	}
	fmt.Println("message: ", &c.Msg)
	return &c, nil
}

func (b *Bithumb) getConfig() (*configResp, error) {
	var c configResp
	err := b.get(b.url+"/spot/config", &c)
	if err != nil {
		return &c, err
	}
	fmt.Println("message: ", &c.Data)
	return &c, nil
}
