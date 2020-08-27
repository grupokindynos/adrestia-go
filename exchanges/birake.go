package exchanges

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/common/hestia"
	"io/ioutil"
	"net/http"
	"strings"
)

type Birake struct {
	Name string
	user string
	authorization string
	client *http.Client
}

func NewBirake(params models.ExchangeParams) *Birake {
	b := new(Birake)
	b.Name = params.Name
	b.user = params.Keys.PublicKey
	b.authorization = params.Keys.PrivateKey
	b.client = &http.Client{}
	return b
}

func (b *Birake) GetName() (string, error) {
	return b.Name, nil
}

type getAddressParams struct {
	Asset string `json:"asset"`
}

type getAddressResponse struct {
	Address string `json:"address"`
}

func (b *Birake) GetAddress(asset string) (string, error) {
	params := getAddressParams{Asset:asset}
	body, _ := json.Marshal(params)

	resBytes, err := b.doRequest("POST", "private/deposit", body)
	if err != nil {
		return "", errors.New("Birake::GetAddress::doRequest::" + err.Error())
	}

	res := getAddressResponse{}
	err = json.Unmarshal(resBytes, &res)
	if err != nil {
		return "", errors.New("Birake::GetAddress::UnmarshalResponse::" + err.Error())
	}

	return res.Address, nil
}

type getBalanceResponse []struct {
	Name  string     `json:"name"`
	Free  float64    `json:"free"`
	Used  float64    `json:"used"`
	Total float64    `json:"total"`
}

func (b *Birake) GetBalance(asset string) (float64, error) {
	resBytes, err := b.doRequest("POST", "private/balances", []byte{})
	if err != nil {
		return 0.0, errors.New("Birake::GetBalance::doRequest::" + err.Error())
	}

	res := getBalanceResponse{}
	err = json.Unmarshal(resBytes, &res)
	if err != nil {
		return 0.0, errors.New("Birake::GetBalance::Unmarshal::" + err.Error())
	}

	for _, bal := range res {
		if bal.Name == asset {
			return bal.Free, nil
		}
	}

	return 0.0, errors.New("Asset " + asset + " not found")
}

type sellAtMarketPriceParams struct {
	Amount float64	`json:"amount"`
	Price  float64	`json:"price"`
	Type   string	`json:"type"`
	Market string	`json:"market"`
}

func (b *Birake) SellAtMarketPrice(order hestia.Trade) (string, error) {
	params := sellAtMarketPriceParams{
		Amount: order.Amount,
		Price:  0, // create function to calculate price
		Type:   order.Side,
		Market: order.Symbol,
	}
	body, _ := json.Marshal(params)

	resBytes, err := b.doRequest("POST", "private/addOrder", body)
	if err != nil {
		return "", errors.New("Birake::SellAtMarketPrice::doRequest::" + err.Error())
	}

	return string(resBytes), nil // missing check what it returns
}

type withdrawParams struct {
	Asset string `json:"asset"`
	Address string `json:"address"`
	Amount float64 `json:"amount"`
}

type withdrawResponse struct {
	Id string `json:"id"`
	Amount float64 `json:"amount"`
}

func (b *Birake) Withdraw(asset string, address string, amount float64) (string, error) {
	params := withdrawParams{
		Asset:   asset,
		Address: address,
		Amount:  amount,
	}

	body, _ := json.Marshal(params)

	resBytes, err := b.doRequest("POST", "private/withdraw", body)
	if err != nil {
		return "", errors.New("Birake::Withdraw::doRequest::" + err.Error())
	}

	res := withdrawResponse{}
	err = json.Unmarshal(resBytes, &res)
	if err != nil {
		return "", errors.New("Birake::Withdraw::Unmarshal::" + err.Error())
	}

	return res.Id, nil
}

type ordersParams struct {
	Pair  string `json:"pair"`
	Limit int    `json:"limit"`
}

type ordersResponse []struct {
	ID            string  `json:"id"`
	Pair          string  `json:"pair"`
	Price         float64 `json:"price"`
	InitialAmount float64 `json:"initialAmount"`
	Amount        float64 `json:"amount"`
	Side          string  `json:"side"`
	Type          string  `json:"type"`
	Timestamp     string  `json:"timestamp"`
	Status        string  `json:"status"`
}

func (b *Birake) GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error) {
	params := ordersParams{
		Pair: order.Symbol,
	}
	body, _ := json.Marshal(params)

	resBytes, err := b.doRequest("POST", "/private/openOrders", body)
	if err != nil {
		return hestia.ExchangeOrderInfo{}, errors.New("Birake::GetOrderStatus::doRequest::" + err.Error())
	}

	res := ordersResponse{}
	err = json.Unmarshal(resBytes, &res)
	if err != nil {
		return hestia.ExchangeOrderInfo{}, errors.New("Birake::GetOrderStatus::Unmarshal::" + err.Error())
	}

	for _, openOrder := range res {
		if openOrder.ID == order.OrderId {
			status := hestia.ExchangeOrderStatusCompleted
			if openOrder.Status == "open" {
				status = hestia.ExchangeOrderStatusOpen
			}

			return hestia.ExchangeOrderInfo{
				Status:         status,
				ReceivedAmount: openOrder.Amount,
			}, nil
		}
	}

	return hestia.ExchangeOrderInfo{}, errors.New("Birake::GetOrderStatus::OrderId not found")
}

type getPairResponse[] struct {
	Symbol 		string `json:"symbol"`
	Quote 		string `json:"quote"`
	Base 		string `json:"base"`
	TickSize 	float64 `json:"tickSize"`
	MinPrice 	float64 `json:"minPrice"`
	MinVolume 	float64 `json:"minVolume"`
}

func (b *Birake) GetPair(fromCoin string, toCoin string) (models.TradeInfo, error) {
	resBytes, err := b.doRequest("GET", "public/markets", []byte{})
	if err != nil {
		return models.TradeInfo{}, errors.New("Birake:::GetPair::doRequest::" + err.Error())
	}

	pairs := getPairResponse{}
	err = json.Unmarshal(resBytes, &pairs)

	order := models.TradeInfo{}

	for _, pair := range pairs {
		if strings.Contains(pair.Symbol, fromCoin) && strings.Contains(pair.Symbol, toCoin) {
			order.Book = pair.Symbol
			if pair.Base == fromCoin {
				order.Type = "sell"
			} else {
				order.Type = "buy"
			}

			return order, nil
		}
	}

	return order, errors.New("Birake::GetPair::Symbol not found")
}

func (b *Birake) GetWithdrawalTxHash(txId string, asset string) (string, error) {
	return "", nil
}

func (b *Birake) GetDepositStatus(addr string, txId string, asset string) (hestia.ExchangeOrderInfo, error) {
	return hestia.ExchangeOrderInfo{}, nil
}

func (b *Birake) doRequest(method string, path string, body []byte) ([]byte, error) {
	buf := bytes.NewBuffer(body)

	req, err := http.NewRequest(method, "https://api.birake.com/v5/"+path, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("birake-user", b.user)
	req.Header.Set("birake-authorization", b.authorization)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 400:
	case 401:
	case 403:
	case 404:
	case 405:
	case 429:
		errString, err := getError(respBody)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error code %d: %s", resp.StatusCode, errString)
	case 503:
	case 500:
		return nil, fmt.Errorf("server error: %s", string(body))
	default:
	}

	return respBody, nil
}