package exchanges

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models"
	coinfactory "github.com/grupokindynos/common/coin-factory"
	"github.com/grupokindynos/common/hestia"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Stex struct {
	exchangeInfo hestia.ExchangeInfo
	client      *http.Client
	currencyIDs map[string]currencyInfo
	pairIDs     map[string]pairInfo
}

type stexResponseCurrencies struct {
	Success bool `json:"success"`
	Data    []struct {
		ID        int    `json:"id"`
		Code      string `json:"code"`
		Precision int    `json:"precision"`
		MinimumTxConfirmations int `json:"minimum_tx_confirmations"`
		ProtocolSpecificSettings []struct {
			Name string `json:"protocol_name"`
			Id int `json:"protocol_id"`
		} `json:"protocol_specific_settings"`
	} `json:"data"`
}

type stexResponsePairs struct {
	Success bool `json:"success"`
	Data    []struct {
		ID              int    `json:"id"`
		Symbol          string `json:"symbol"`
		BasePrecision   int32  `json:"currency_precision"`
		BaseCurrency    string `json:"currency_code"`
		MarketPrecision int32  `json:"market_precision"`
		MarketCurrency  string `json:"market_code"`
	} `json:"data"`
}

type currencyInfo struct {
	id        int
	code      string
	precision int
	minimumConfirmations int
	protocolIds map[string]int
}

type pairInfo struct {
	id              int
	basePrecision   int32
	base string
	marketPrecision int32
	market string
}

// NewStex creates a new STEX exchange instance.
func NewStex(exchange hestia.ExchangeInfo) (*Stex, error) {
	s := &Stex{
		exchangeInfo: exchange,
		client:      http.DefaultClient,
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
		var minConfirm int
		// USDT code
		if currency.Code == "USDT" {
			minConfirm = 15
		} else {
			minConfirm = currency.MinimumTxConfirmations
		}
		s.currencyIDs[currency.Code] = currencyInfo{
			id:        currency.ID,
			code:      currency.Code,
			precision: currency.Precision,
			minimumConfirmations: minConfirm,
			protocolIds: make(map[string]int),
		}
		for _, protocol := range currency.ProtocolSpecificSettings {
			s.currencyIDs[currency.Code].protocolIds[protocol.Name] = protocol.Id
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

func (s *Stex) GetBalance(coin string) (float64, error) {
	out, err := s.doRequest("GET", "/profile/wallets", nil)
	if err != nil {
		return 0, err
	}

	var stexBalances stexResponseBalances
	if err := json.Unmarshal(out, &stexBalances); err != nil {
		return 0, err
	}

	if !stexBalances.Success {
		return 0, errors.New("retrieving balances unsuccessful")
	}

	for _, b := range stexBalances.Data {
		if coin == b.Currency{
			val, _ := b.Balance.Float64()
			return val, nil
		}
	}

	return 0, errors.New("coin not found")
}

type stexResponseTicker struct {
	Success bool `json:"success"`
	Data    struct {
		Symbol string          `json:"symbol"`
		Ask    decimal.Decimal `json:"ask"`
		Bid    decimal.Decimal `json:"bid"`
		Last   decimal.Decimal `json:"last"`
		Low    decimal.Decimal `json:"low"`
		High   decimal.Decimal `json:"high"`
	} `json:"data"`
}

func (s *Stex) getMarketPrice(pair string, side string) (*decimal.Decimal, error) {
	tickerBytes, err := s.doRequest("GET", fmt.Sprintf("/public/ticker/%s", pair), nil)
	if err != nil {
		return nil, err
	}

	var ticker stexResponseTicker
	if err := json.Unmarshal(tickerBytes, &ticker); err != nil {
		return nil, err
	}

	if side == "buy" {
		return &ticker.Data.High, nil
	} else {
		return &ticker.Data.Low, nil
	}
}

type stexOrder struct {
	CurrencyPairID   int     `json:"currency_pair_id"`
	Amount           string  `json:"amount"`
	Price            string  `json:"price"`
	Amount2          string  `json:"amount2"`
	Count            int     `json:"count"`
	CumulativeAmount float64 `json:"cumulative_amount"`
	Type             string  `json:"type,omitempty"`
}

type stexOrderBookResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Ask []stexOrder `json:"ask"`
		Bid []stexOrder `json:"bid"`
	} `json:"data"`
}

func (s *Stex) getBestPrice(amount decimal.Decimal, market string, side string) (decimal.Decimal, error) {
	respBytes, err := s.doRequest("GET", fmt.Sprintf("/public/orderbook/%d", s.pairIDs[market].id), nil)
	if err != nil {
		return decimal.Decimal{}, err
	}
	var res stexOrderBookResponse
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return decimal.Decimal{}, err
	}
	var orders []stexOrder
	if side == "buy" {
		orders = res.Data.Ask
	} else {
		orders = res.Data.Bid
	}

	cumulativeAmount := decimal.NewFromFloat(0)
	var price decimal.Decimal
	for _, order := range orders {
		cumulativeAmount = cumulativeAmount.Add(decimal.NewFromFloat(order.CumulativeAmount))
		if cumulativeAmount.GreaterThan(amount) {
			price, err = decimal.NewFromString(order.Price)
			if err != nil {
				return price, err
			}
			break
		}
	}
	return price, nil
}

type stexResponseOrder struct {
	Success bool `json:"success"`
	Data    struct {
		ID int `json:"id"`
	} `json:"data"`
}

func (s *Stex) SellAtMarketPrice(sellOrder hestia.Trade) (string, error) {
	amount := decimal.NewFromFloat(sellOrder.Amount)

	marketPair := sellOrder.Symbol

	pairInfo := s.pairIDs[marketPair]

	var orderBytes []byte

	if sellOrder.Side == "buy" {
		price, err := s.getMarketPrice(fmt.Sprintf("%d", pairInfo.id), sellOrder.Side)
		if err != nil {
			return "", err
		}
		buyAmount := amount.Div(*price)
		bestPrice, err := s.getBestPrice(buyAmount, marketPair, "buy")
		if err != nil {
			return "", err
		}

		values := url.Values{}
		values.Set("type", "BUY")
		values.Set("amount", buyAmount.StringFixed(pairInfo.marketPrecision))
		values.Set("price", bestPrice.String())

		orderBytes, err = s.doRequest("POST", fmt.Sprintf("/trading/orders/%d", pairInfo.id), values)
		if err != nil {
			return "", err
		}
	} else {
		bestPrice, err := s.getBestPrice(amount, marketPair, "sell")
		if err != nil {
			return "", err
		}

		values := url.Values{}
		values.Set("type", "SELL")
		values.Set("amount", amount.StringFixed(pairInfo.marketPrecision))
		values.Set("price", bestPrice.String())

		orderBytes, err = s.doRequest("POST", fmt.Sprintf("/trading/orders/%d", pairInfo.id), values)
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

func (s *Stex) Withdraw(coin string, address string, amount float64) (string, error) {
	currencyID := s.currencyIDs[strings.ToUpper(coin)]
	amountDec := decimal.NewFromFloat(amount)

	vals := url.Values{}
	vals.Set("currency_id", fmt.Sprintf("%d", currencyID.id))
	vals.Set("amount", amountDec.StringFixed(int32(currencyID.precision)))
	vals.Set("address", address)
	if currencyID.code == "USDT" {
		vals.Set("protocol_id", fmt.Sprintf("%d", currencyID.protocolIds["ERC20"]))
	}

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

type stexResponseOrderStatusEmpty struct {
	Success bool          `json:"success"`
	Data    []interface{} `json:"data"`
}


func (s *Stex) GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error) {
	statusBytes, err := s.doRequest("GET", fmt.Sprintf("/trading/order/%s", order.OrderId), nil)
	if err != nil {
		return hestia.ExchangeOrderInfo{}, err
	}

	log.Println(string(statusBytes))

	var status stexResponseOrderStatus
	var emptyStatus stexResponseOrderStatusEmpty

	if err := json.Unmarshal(statusBytes, &status); err != nil {
		if err = json.Unmarshal(statusBytes, &emptyStatus); err != nil {
			return hestia.ExchangeOrderInfo{}, err
		}
		return hestia.ExchangeOrderInfo{Status: hestia.ExchangeOrderStatusOpen}, nil
	}

	orderStatus := hestia.ExchangeOrderInfo{
		Status:          hestia.ExchangeOrderStatusOpen,
	}

	if status.Data.Status == "FINISHED" {
		orderStatus.Status = hestia.ExchangeOrderStatusCompleted
	}

	orderStatus.ReceivedAmount, _ = status.Data.ProcessedAmount.Float64()

	return orderStatus, nil
}

func (s *Stex) GetPair(fromCoin string, toCoin string) (models.TradeInfo, error) {
	fromUpper := strings.ToUpper(fromCoin)
	toUpper := strings.ToUpper(toCoin)

	var book *pairInfo
	var symbol string
	for key, pair := range s.pairIDs {
		if (fromUpper == pair.market && toUpper == pair.base) || (fromUpper == pair.base && toUpper == pair.market) {
			book = &pair
			symbol = key
			break
		}
	}

	if book == nil {
		return models.TradeInfo{}, fmt.Errorf("could not find instrument for symbols %s and %s", fromCoin, toCoin)
	}

	var orderSide models.TradeInfo
	orderSide.Book = symbol
	if book.market == fromCoin {
		orderSide.Type = "sell"
	} else {
		orderSide.Type = "buy"
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

	log.Println(string(withdrawInfoBytes))

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

func (s *Stex) GetDepositStatus(addr string, txId string, asset string) (hestia.ExchangeOrderInfo, error) {
	coinInfo, _ := coinfactory.GetCoin(asset)
	if coinInfo.Info.Token {
		if val, err := blockbookConfirmed(addr, txId, s.currencyIDs[asset].minimumConfirmations); err == nil {
			return hestia.ExchangeOrderInfo{
				Status: hestia.ExchangeOrderStatusCompleted,
				ReceivedAmount: val,
			}, nil
		} else {
			return hestia.ExchangeOrderInfo{}, err
		}
	}

	depositResponseBytes, err := s.doRequest("GET", "/profile/deposits", nil)
	if err != nil {
		return hestia.ExchangeOrderInfo{}, err
	}

	var depositResponse stexDepositResponse
	if err := json.Unmarshal(depositResponseBytes, &depositResponse); err != nil {
		return hestia.ExchangeOrderInfo{}, err
	}

	for _, d := range depositResponse.Data {
		if d.TxID == txId {
			amount, _ := d.Amount.Float64()
			if d.Status == "Finished" {
				return hestia.ExchangeOrderInfo{
					Status: hestia.ExchangeOrderStatusCompleted,
					ReceivedAmount: amount,
				}, nil
			}
			if d.Status == "Processing" || d.Status == "Checking by System" {
				return hestia.ExchangeOrderInfo {
					Status: hestia.ExchangeOrderStatusCompleted,
					ReceivedAmount: amount,
				}, nil
			}
			return hestia.ExchangeOrderInfo{
				Status: hestia.ExchangeOrderStatusError,
			}, nil
		}
	}

	return hestia.ExchangeOrderInfo{}, errors.New("could not find deposit")
}

type stexWalletResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID              int             `json:"id"`
		MultiDepositAddresses [] struct {
			Address string `json:"address"`
			ProtocolName string `json:"protocol_name"`
		} `json:"multi_deposit_addresses"`
	} `json:"data"`
}

func (s *Stex) GetAddress(asset string) (string, error) {
	coinUpper := strings.ToUpper(asset)
	info := s.currencyIDs[coinUpper]

	walletResponseBytes, err := s.doRequest("POST", fmt.Sprintf("/profile/wallets/%d", info.id), nil)
	if err != nil {
		return "", err
	}

	var walletResponse stexWalletResponse

	if err := json.Unmarshal(walletResponseBytes, &walletResponse); err != nil {
		return "", err
	}

	for _, depositAddress := range walletResponse.Data.MultiDepositAddresses {
		if depositAddress.ProtocolName == "ERC20" {
			return depositAddress.Address, nil
		}
	}

	return "", errors.New("coin not found")
}

func (s *Stex) doRequest(method string, path string, body url.Values) ([]byte, error) {
	if body == nil {
		body = url.Values{}
	}

	req, err := http.NewRequest(method, fmt.Sprintf("https://api3.stex.com%s", path), strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.exchangeInfo.ApiPrivateKey))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	resBody, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return nil, errors.New(string(resBody))
	}

	return resBody, nil
}

func (s *Stex) GetName() (string, error) {
	return s.exchangeInfo.Name, nil
}
