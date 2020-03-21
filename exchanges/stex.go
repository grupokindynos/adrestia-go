package exchanges

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/grupokindynos/common/coin-factory/coins"
)

type Stex struct {
	client      *http.Client
	apiKey      string
	currencyIDs map[string]currencyInfo
	pairIDs     map[string]pairInfo
	Obol        obol.ObolService
}

type stexResponseCurrencies struct {
	Success bool `json:"success"`
	Data    []struct {
		ID        int    `json:"id"`
		Code      string `json:"code"`
		Precision int    `json:"precision"`
	} `json:"data"`
}

type stexResponsePairs struct {
	Success bool `json:"success"`
	Data    []struct {
		ID              int    `json:"id"`
		Symbol          string `json:"symbol"`
		BasePrecision   int32  `json:"currency_precision"`
		BaseCurrency string `json:"currency_code"`
		MarketPrecision int32  `json:"market_precision"`
		MarketCurrency string `json:"market_code"`
	} `json:"data"`
}

type currencyInfo struct {
	id        int
	precision int
}

type pairInfo struct {
	id              int
	basePrecision   int32
	base string
	marketPrecision int32
	market string
}

// NewStex creates a new STEX exchange instance.
func NewStex(params Params) (*Stex, error) {
	s := &Stex{
		client:      http.DefaultClient,
		apiKey:      os.Getenv("STEX_API_KEY"),
		Obol:        params.Obol,
		currencyIDs: map[string]currencyInfo{},
		pairIDs:     map[string]pairInfo{},
	}

	currenciesBytes, err := s.doRequest("GET", "/public/currencies", nil)
	if err != nil {
		return nil, err
	}

	var currencies stexResponseCurrencies

	if err := json.Unmarshal(currenciesBytes, &currencies); err != nil {
		return nil, err
	}

	for _, currency := range currencies.Data {
		s.currencyIDs[currency.Code] = currencyInfo{
			id:        currency.ID,
			precision: currency.Precision,
		}
	}

	pairsBytes, err := s.doRequest("GET", "/public/currency_pairs/list/ALL", nil)
	if err != nil {
		return nil, err
	}

	var pairs stexResponsePairs
	if err := json.Unmarshal(pairsBytes, &pairs); err != nil {
		return nil, err
	}

	for _, pair := range pairs.Data {
		s.pairIDs[pair.Symbol] = pairInfo{
			id:              pair.ID,
			basePrecision:   pair.BasePrecision,
			base: pair.BaseCurrency,
			marketPrecision: pair.MarketPrecision,
			market: pair.MarketCurrency,
		}
	}

	return s, nil
}

type stexResponseBalances struct {
	Success bool `json:"success"`
	Data    []struct {
		ID            int                        `json:"id"`
		Currency      string                     `json:"currency_code"`
		Rates         map[string]decimal.Decimal `json:"rates"`
		Balance       decimal.Decimal            `json:"balance"`
		FrozenBalance decimal.Decimal            `json:"frozen_balance"`
	} `json:"data"`
}

func (s *Stex) GetBalances() ([]balance.Balance, error) {
	out, err := s.doRequest("GET", "/profile/wallets", nil)
	if err != nil {
		return nil, err
	}

	var stexBalances stexResponseBalances
	if err := json.Unmarshal(out, &stexBalances); err != nil {
		return nil, err
	}

	if !stexBalances.Success {
		return nil, errors.New("retrieving balances unsuccessful")
	}

	balances := make([]balance.Balance, 0, len(stexBalances.Data))
	for _, b := range stexBalances.Data {
		confirmedBalanceFloat, _ := b.Balance.Float64()
		rate, found := b.Rates["BTC"]
		if !found {
			rateFloat, _ := s.Obol.GetCoin2CoinRates("BTC", b.Currency)
			rate = decimal.NewFromFloat(rateFloat)
		}
		rateFloat, _ := rate.Float64()
		possibleBalance := balance.Balance{
			Ticker:             b.Currency,
			ConfirmedBalance:   confirmedBalanceFloat,
			UnconfirmedBalance: 0,
			RateBTC:            rateFloat,
			DiffBTC:            0.0,
			IsBalanced:         false,
		}

		if possibleBalance.GetTotalBalance() > 0 {
			balances = append(balances, possibleBalance)
		}
	}

	return balances, nil
}

type stexResponseTicker struct {
	Success bool `json:"success"`
	Data    struct {
		Symbol string          `json:"symbol"`
		Ask    decimal.Decimal `json:"ask"`
		Bid    decimal.Decimal `json:"bid"`
		Last   decimal.Decimal `json:"last"`
	} `json:"data"`
}

func (s *Stex) getMarketPrice(pair string) (*decimal.Decimal, error) {
	tickerBytes, err := s.doRequest("GET", fmt.Sprintf("/public/ticker/%s", pair), nil)
	if err != nil {
		return nil, err
	}

	var ticker stexResponseTicker
	if err := json.Unmarshal(tickerBytes, &ticker); err != nil {
		return nil, err
	}

	return &ticker.Data.Bid, nil
}

type stexResponseOrder struct {
	Success bool `json:"success"`
	Data    struct {
		ID int `json:"id"`
	} `json:"data"`
}

func (s *Stex) SellAtMarketPrice(sellOrder hestia.ExchangeOrder) (string, error) {
	market, base := sellOrder.GetTradingPair()
	amount := decimal.NewFromFloat(sellOrder.Amount)

	marketPair := fmt.Sprintf("%s_%s", strings.ToUpper(market), strings.ToUpper(base))

	pairInfo := s.pairIDs[marketPair]

	var orderBytes []byte

	if sellOrder.Side == "buy" {
		price, err := s.getMarketPrice(marketPair)
		if err != nil {
			return "", err
		}

		buyAmount := amount.Div(*price)

		values := url.Values{}
		values.Set("type", "BUY")
		values.Set("amount", buyAmount.StringFixed(pairInfo.marketPrecision))
		values.Set("price", price.String())

		orderBytes, err = s.doRequest("POST", fmt.Sprintf("/trading/orders/%d", pairInfo.id), nil)
		if err != nil {
			return "", err
		}
	} else {
		price, err := s.getMarketPrice(marketPair)
		if err != nil {
			return "", err
		}

		values := url.Values{}
		values.Set("type", "BUY")
		values.Set("amount", amount.StringFixed(pairInfo.marketPrecision))
		values.Set("price", price.String())

		orderBytes, err = s.doRequest("POST", fmt.Sprintf("/trading/orders/%d", pairInfo.id), nil)
		if err != nil {
			return "", err
		}
	}

	var order stexResponseOrder
	if err := json.Unmarshal(orderBytes, &order); err != nil {
		return "", err
	}

	if !order.Success {
		return "", errors.New("order unsuccessful")
	}

	return fmt.Sprintf("%d", order.Data.ID), nil
}

type stexWithdrawResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID int `json:"id"`
	} `json:"data"`
}

func (s *Stex) Withdraw(coin coins.Coin, address string, amount float64) (string, error) {
	currencyID := s.currencyIDs[strings.ToUpper(coin.Info.Tag)]
	amountDec := decimal.NewFromFloat(amount)

	vals := url.Values{}
	vals.Set("currency_id", fmt.Sprintf("%d", currencyID.id))
	vals.Set("amount", amountDec.StringFixed(int32(currencyID.precision)))
	vals.Set("address", address)

	withdrawResponseBytes, err := s.doRequest("POST", "/profile/withdraw", vals)
	if err != nil {
		return "", err
	}

	var withdraw stexWithdrawResponse

	if err := json.Unmarshal(withdrawResponseBytes, &withdraw); err != nil {
		return "", err
	}

	if !withdraw.Success {
		return "", fmt.Errorf("withdraw unsuccessful")
	}

	return fmt.Sprintf("%d", withdraw.Data.ID), nil
}

type stexResponseOrderStatus struct {
	Success bool `json:"success"`
	Data    struct {
		ID              int             `json:"id"`
		Status          string          `json:"status"`
		ProcessedAmount decimal.Decimal `json:"processed_amount"`
		InitialAmount   decimal.Decimal `json:"initial_amount"`
	} `json:"data"`
}

func (s *Stex) GetOrderStatus(order hestia.ExchangeOrder) (hestia.OrderStatus, error) {
	statusBytes, err := s.doRequest("GET", fmt.Sprintf("/trading/orders/%s", order.OrderId), nil)
	if err != nil {
		return hestia.OrderStatus{}, err
	}

	var status stexResponseOrderStatus

	if err := json.Unmarshal(statusBytes, &status); err != nil {
		return hestia.OrderStatus{}, err
	}

	orderStatus := hestia.OrderStatus{
		Status:          hestia.ExchangeStatusOpen,
		AvailableAmount: 0,
	}

	if status.Data.Status == "FINISHED" {
		orderStatus.Status = hestia.ExchangeStatusCompleted
	}

	orderStatus.AvailableAmount, _ = status.Data.ProcessedAmount.Float64()

	return orderStatus, nil
}

func (s *Stex) GetPair(fromCoin string, toCoin string) (OrderSide, error) {
	fromUpper := strings.ToUpper(fromCoin)
	toUpper := strings.ToUpper(toCoin)

	var book *pairInfo
	for _, pair := range s.pairIDs {
		if (fromUpper == pair.market && toUpper == pair.base) || (fromUpper == pair.base && toUpper == pair.market) {
			book = &pair
			break
		}
	}

	if book == nil {
		return OrderSide{}, fmt.Errorf("could not find instrument for symbols %s and %s", fromCoin, toCoin)
	}

	var orderSide OrderSide
	orderSide.Book = book.market + book.base
	if book.market == fromCoin {
		orderSide.Type = "sell"
		orderSide.ReceivedCurrency = book.base
		orderSide.SoldCurrency = book.market
	} else {
		orderSide.Type = "buy"
		orderSide.ReceivedCurrency = book.market
		orderSide.SoldCurrency = book.base
	}

	return orderSide, nil
}

type stexWithdrawInfoResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID              int             `json:"id"`
		TxID string `json:"txid"`
	} `json:"data"`
}

func (s *Stex) GetWithdrawalTxHash(txId string, asset string) (string, error) {
	withdrawInfoBytes, err := s.doRequest("GET", fmt.Sprintf("/profile/withdrawals/%s", txId), nil)
	if err != nil {
		return "", err
	}

	var withdrawInfo stexWithdrawInfoResponse
	if err := json.Unmarshal(withdrawInfoBytes, &withdrawInfo); err != nil {
		return "", err
	}

	return withdrawInfo.Data.TxID, nil
}

type stexDepositResponse struct {
	Success bool `json:"success"`
	Data    []struct {
		ID              int             `json:"id"`
		TxID string `json:"txid"`
		Status string `json:"status"`
		Amount decimal.Decimal `json:"amount"`
	} `json:"data"`
}

func (s *Stex) GetDepositStatus(txid string, _ string) (hestia.OrderStatus, error) {
	depositResponseBytes, err := s.doRequest("GET", "/profile/deposits", nil)
	if err != nil {
		return hestia.OrderStatus{}, err
	}

	var depositResponse stexDepositResponse
	if err := json.Unmarshal(depositResponseBytes, &depositResponse); err != nil {
		return hestia.OrderStatus{}, err
	}

	for _, d := range depositResponse.Data {
		if d.TxID == txid {
			amount, _ := d.Amount.Float64()
			if d.Status == "Finished" {
				return hestia.OrderStatus{
					Status: hestia.ExchangeStatusCompleted,
					AvailableAmount: amount,
				}, nil
			}
			if d.Status == "Processing" || d.Status == "Checking by System" {
				return hestia.OrderStatus{
					Status: hestia.ExchangeStatusCompleted,
					AvailableAmount: amount,
				}, nil
			}
			return hestia.OrderStatus{
				Status: hestia.ExchangeStatusError,
			}, nil
		}
	}

	return hestia.OrderStatus{}, errors.New("could not find deposit")
}

type stexWalletResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID              int             `json:"id"`
		DepositAddress struct {
			Address string `json:"address"`
		} `json:"deposit_address"`
	} `json:"data"`
}

func (s *Stex) GetAddress(coin coins.Coin) (string, error) {
	coinUpper := strings.ToUpper(coin.Info.Tag)
	info := s.currencyIDs[coinUpper]

	walletResponseBytes, err := s.doRequest("POST", fmt.Sprintf("/profile/wallets/%d", info.id), nil)
	if err != nil {
		return "", err
	}

	var walletResponse stexWalletResponse

	if err := json.Unmarshal(walletResponseBytes, &walletResponse); err != nil {
		return "", err
	}

	return walletResponse.Data.DepositAddress.Address, nil
}

func (s *Stex) doRequest(method string, path string, body url.Values) ([]byte, error) {
	if body == nil {
		body = url.Values{}
	}

	req, err := http.NewRequest(method, fmt.Sprintf("https://api3.stex.com%s", path), strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("got code: %d", res.StatusCode)
	}

	return ioutil.ReadAll(res.Body)
}

func (s *Stex) GetName() (string, error) {
	return "STEX", nil
}

var _ IExchange = &Stex{}
