package exchanges

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

	return 0, nil
}

func (b *Bithumb) SellAtMarketPrice(order hestia.Trade) (string, error) {
	return "", nil
}

func (b *Bithumb) Withdraw(asset string, address string, amount float64) (string, error) {
	return "", nil
}

func (b *Bithumb) GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error) {
	h := hestia.ExchangeOrderInfo{}
	return h, nil
}

func (b *Bithumb) GetPair(fromCoin string, toCoin string) (models.TradeInfo, error) {
	t := models.TradeInfo{}
	return t, nil
}

func (b *Bithumb) GetWithdrawalTxHash(txId string, asset string) (string, error) {
	return "", nil
}

//  Gets the deposit status from an asset's exchange.
func (b *Bithumb) GetDepositStatus(addr string, txId string, asset string) (hestia.ExchangeOrderInfo, error) {
	e := hestia.ExchangeOrderInfo{}
	return e, nil
}

// Functions to get Bithumb API data
type assetsResp struct {
	baseResp
	Data []struct {
		CoinType    string
		Count       string
		Frozen      string
		Type        string
		BtcQuantity string
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

func (b *Bithumb) post(url string, params interface{}, result interface{}) {
	preMap := b.struct2map(params)
	preMap["apiKey"] = b.user
	preMap["timestamp"] = strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	preMap["signature"] = b.sign(preMap)
	err := post(url, preMap, result)
	handleErr(err)
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

func (b *Bithumb) Assets(coinType string) *assetsResp {
	var r assetsResp
	p := struct {
		CoinType  string `json:"coinType"`
		AssetType string `json:"assetType"`
	}{
		coinType, "spot",
	}
	b.post(b.url+"/spot/assetList", p, &r)
	return &r
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
