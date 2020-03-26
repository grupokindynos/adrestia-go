package exchanges

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models"
	"github.com/grupokindynos/common/hestia"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Crex24 struct {
	exchangeInfo hestia.ExchangeInfo
	client *http.Client
}

func NewCrex24(exchange hestia.ExchangeInfo) *Crex24 {
	return &Crex24{
		exchangeInfo: exchange,
		client: &http.Client{},
	}
}

func (c *Crex24) GetName() (string, error) {
	return "Crex24", nil
}

type errorDescriptionResponse struct {
	ErrorDescription string `json:"errorDescription"`
}

func getError(body []byte) (string, error) {
	var res errorDescriptionResponse
	err := json.Unmarshal(body, &res)
	if err != nil {
		return "", err
	}
	return res.ErrorDescription, nil
}

func (c *Crex24) doRequest(method string, path string, body []byte) ([]byte, error) {
	buf := bytes.NewBuffer(body)

	req, err := http.NewRequest(method, "https://api.crex24.com"+path, buf)
	if err != nil {
		return nil, err
	}

	hmacB64, err := base64.StdEncoding.DecodeString(c.exchangeInfo.ApiPrivateKey)
	if err != nil {
		return nil, err
	}

	nonce := fmt.Sprintf("%d", time.Now().Unix())

	h := hmac.New(sha512.New, hmacB64)
	h.Write([]byte(path))
	h.Write([]byte(nonce))
	h.Write(body)

	sig := h.Sum(nil)

	req.Header.Set("x-crex24-api-key", c.exchangeInfo.ApiPublicKey)
	req.Header.Set("x-crex24-api-nonce", nonce)
	req.Header.Set("x-crex24-api-sign", base64.StdEncoding.EncodeToString(sig))

	fmt.Println(req.Header)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(respBody))

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

type addressResponse struct {
	Currency  string `json:"currency"`
	Address   string `json:"address"`
	PaymentID string `json:"paymentId"`
}

func (c *Crex24) GetAddress(coin string) (string, error) {
	out, err := c.doRequest("GET", fmt.Sprintf("/v2/account/depositAddress?currency=%s", coin), []byte{})
	if err != nil {
		return "", err
	}

	var addrRes addressResponse
	if err := json.Unmarshal(out, &addrRes); err != nil {
		return "", err
	}

	return addrRes.Address, nil
}

type crex24Balance struct {
	Currency  string          `json:"currency"`
	Available decimal.Decimal `json:"available"`
	Reserved  decimal.Decimal `json:"reserved"`
}

func (c *Crex24) GetBalance(coin string) (float64, error) {
	out, err := c.doRequest("GET", "/v2/account/balance", []byte{})
	if err != nil {
		return 0, err
	}

	var crexBalances []crex24Balance
	if err := json.Unmarshal(out, &crexBalances); err != nil {
		return 0, err
	}

	for _, asset := range crexBalances {
		if coin == asset.Currency {
			confirmed, _ := asset.Available.Float64()
			return confirmed, nil
		}
	}

	return 0, errors.New("coin not found")
}

type crex24SellRequest struct {
	Instrument string          `json:"instrument"`
	Side       string          `json:"side"`
	Volume     decimal.Decimal `json:"volume"`
	Type       string          `json:"type"`
	Price      decimal.Decimal `json:"price"`
}

type crex24PriceResponse struct {
	Instrument string `json:"instrument"`
	Last decimal.Decimal `json:"last"`
	Bid decimal.Decimal `json:"bid"`
	Ask decimal.Decimal `json:"ask"`
}

func (c *Crex24) getMarketPrice(market string) (out decimal.Decimal, err error) {
	resBytes, err := c.doRequest("GET", "/v2/public/tickers", []byte{})
	if err != nil {
		return
	}

	var prices []crex24PriceResponse
	err = json.Unmarshal(resBytes, &prices)
	if err != nil {
		return
	}

	marketLower := strings.ToLower(market)
	for _, p := range prices {
		if strings.ToLower(p.Instrument) == marketLower {
			return p.Bid, nil
		}
	}

	err = fmt.Errorf("could not find instrument %s", market)
	return
}

type crex24IDResponse struct {
	ID int `json:"id"`
}

func (c *Crex24) SellAtMarketPrice(sellOrder hestia.Trade) (string, error) {
	market, base := sellOrder.GetTradingPair()
	amount := decimal.NewFromFloat(sellOrder.Amount)

	var resBytes []byte

	if sellOrder.Side == "buy" {
		name := strings.ToUpper(getMarketName(base, market))
		price, err := c.getMarketPrice(name)
		if err != nil {
			return "", err
		}
		buyAmount := amount.Div(price)

		req := crex24SellRequest{
			Instrument: name,
			Side:       "buy",
			Volume:     buyAmount,
			Type:       "limit",
			Price:      price,
		}

		reqBytes, err := json.Marshal(req)
		if err != nil {
			return "", err
		}

		resBytes, err = c.doRequest("POST", "/v2/trading/placeOrder", reqBytes)
		if err != nil {
			return "", err
		}
	} else {
		name := strings.ToUpper(getMarketName(base, market))

		req := crex24SellRequest{
			Instrument: name,
			Side:       "buy",
			Volume:     amount,
			Type:       "limit",
			Price:      decimal.Zero,
		}

		reqBytes, err := json.Marshal(req)
		if err != nil {
			return "", err
		}

		resBytes, err = c.doRequest("POST", "/v2/trading/placeOrder", reqBytes)
		if err != nil {
			return "", err
		}
	}

	var res crex24IDResponse
	if err := json.Unmarshal(resBytes, &res); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", res.ID), nil
}

type crex24WithdrawRequest struct {
	Currency string `json:"currency"`
	Amount decimal.Decimal `json:"amount"`
	Address string `json:"address"`
}

func (c *Crex24) Withdraw(coin string, address string, amount float64) (string, error) {
	amountDec := decimal.NewFromFloat(amount)
	req := crex24WithdrawRequest{
		Currency: coin,
		Amount:   amountDec,
		Address:  address,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resBytes, err := c.doRequest("POST", "/v2/account/withdraw", reqBytes)
	if err != nil {
		return "", err
	}

	var res crex24IDResponse
	if err := json.Unmarshal(resBytes, &res); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", res.ID), nil
}

type crex24OrderStatus struct {
	Volume decimal.Decimal `json:"volume"`
	RemainingVolume decimal.Decimal `json:"remainingVolume"`
	Status string `json:"status"`
}

func (c *Crex24) GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error) {
	resBytes, err := c.doRequest("GET", fmt.Sprintf("/v2/trading/orderStatus?id=%s", order.OrderId), []byte{})
	if err != nil {
		return hestia.ExchangeOrderInfo{}, err
	}

	var orderStatus crex24OrderStatus
	if err := json.Unmarshal(resBytes, &orderStatus); err != nil {
		return hestia.ExchangeOrderInfo{}, err
	}

	status := hestia.ExchangeOrderInfo{}
	status.Status = hestia.ExchangeOrderStatusError

	switch orderStatus.Status {
	case "submitting":
	case "unfilledActive":
		status.Status = hestia.ExchangeOrderStatusOpen
	case "partiallyFilledActive":
		status.Status = hestia.ExchangeOrderStatusOpen
		availableFloat, _ := orderStatus.Volume.Sub(orderStatus.RemainingVolume).Float64()
		status.ReceivedAmount = availableFloat
	case "filled":
		status.Status = hestia.ExchangeOrderStatusCompleted
	}

	return status, nil
}

type crex24Instrument struct {
	Symbol string `json:"symbol"`
	BaseCurrency string `json:"baseCurrency"`
	QuoteCurrency string `json:"quoteCurrency"`
}

func (c *Crex24) GetPair(fromCoin string, toCoin string) (models.TradeInfo, error) {
	fromCoin = strings.ToUpper(fromCoin)
	toCoin = strings.ToUpper(toCoin)

	respBytes, err := c.doRequest("GET", "/v2/public/instruments", []byte{})
	if err != nil {
		return models.TradeInfo{}, err
	}

	var instruments []crex24Instrument
	if err := json.Unmarshal(respBytes, &instruments); err != nil {
		return models.TradeInfo{}, err
	}

	var book *crex24Instrument
	for _, i := range instruments {
		if (i.BaseCurrency == fromCoin && i.QuoteCurrency == toCoin) || (i.QuoteCurrency == fromCoin && i.BaseCurrency == toCoin) {
			book = &i
			break
		}
	}

	if book == nil {
		return models.TradeInfo{}, fmt.Errorf("could not find instrument for symbols %s and %s", fromCoin, toCoin)
	}

	var orderSide models.TradeInfo
	orderSide.Book = book.QuoteCurrency + book.BaseCurrency
	if book.QuoteCurrency == fromCoin {
		orderSide.Type = "sell"
	} else {
		orderSide.Type = "buy"
	}

	return orderSide, nil
}

type crex24WithdrawalStatus struct {
	Type string `json:"type"`
	TxID string `json:"txId"`
}

func (c *Crex24) GetWithdrawalTxHash(txId string, _ string) (string, error) {
	respBytes, err := c.doRequest("GET", fmt.Sprintf("/v2/account/moneyTransferStatus?id=%s", txId), []byte{})
	if err != nil {
		return "", err
	}

	var status []crex24WithdrawalStatus
	if err := json.Unmarshal(respBytes, &status); err != nil {
		return "", err
	}

	if len(status) < 1 {
		return "", fmt.Errorf("could not find withdrawal with id: %s", txId)
	}

	return status[0].TxID, nil
}

type crex24DepositStatus struct {
	TxID string `json:"txId"`
	ConfirmationsRequired int `json:"confirmationsRequired"`
	Confirmations int `json:"confirmationCount"`
	Status string `json:"string"`
	Amount decimal.Decimal `json:"amount"`
}

func (c *Crex24) GetDepositStatus(_ string, asset string) (hestia.ExchangeOrderInfo, error) {
	respBytes, err := c.doRequest("GET", fmt.Sprintf("/v2/account/moneyTransfers?type=deposit&currenct=%s", strings.ToUpper(asset)), []byte{})
	if err != nil {
		return hestia.ExchangeOrderInfo{}, err
	}

	var depositStatus crex24DepositStatus
	if err := json.Unmarshal(respBytes, &depositStatus); err != nil {
		return hestia.ExchangeOrderInfo{}, err
	}

	if depositStatus.Status == "pending" {
		return hestia.ExchangeOrderInfo{
			Status:          hestia.ExchangeOrderStatusOpen,
		}, nil
	} else if depositStatus.Status == "success" {
		amount, _ := depositStatus.Amount.Float64()
		return hestia.ExchangeOrderInfo{
			Status:          hestia.ExchangeOrderStatusCompleted,
			ReceivedAmount: amount,
		}, nil
	}
	return hestia.ExchangeOrderInfo{
		Status:          hestia.ExchangeOrderStatusError,
	}, nil
}