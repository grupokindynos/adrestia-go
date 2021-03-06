package exchanges

import (
	"context"
	"errors"
	"fmt"
	"github.com/grupokindynos/adrestia-go/models"
	l "log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/go-binance"
)

type binanceCurrencyInfo struct {
	withdrawalPrecision int
}

type Binance struct {
	Name string
	binanceApi  binance.Binance
	currenciesInfo map[string]binanceCurrencyInfo
}

func NewBinance(params models.ExchangeParams) *Binance {
	c := new(Binance)
	c.Name = params.Name
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	hmacSigner := &binance.HmacSigner{
		Key: []byte(params.Keys.PrivateKey),
	}
	ctx, _ := context.WithCancel(context.Background())
	binanceService := binance.NewAPIService(
		"https://api.binance.com",
		params.Keys.PublicKey,
		hmacSigner,
		logger,
		ctx,
	)
	c.binanceApi = binance.NewBinance(binanceService)
	//hardcoded withdrawal precision. Change when there's a better solution
	c.currenciesInfo = make(map[string]binanceCurrencyInfo)
	c.currenciesInfo["USDC"] = binanceCurrencyInfo{withdrawalPrecision:6}
	c.currenciesInfo["USDT"] = binanceCurrencyInfo{withdrawalPrecision:6}
	c.currenciesInfo["TUSD"] = binanceCurrencyInfo{withdrawalPrecision:8}
	c.currenciesInfo["BTC"] = binanceCurrencyInfo{withdrawalPrecision:8}
	c.currenciesInfo["DASH"] = binanceCurrencyInfo{withdrawalPrecision:8}
	c.currenciesInfo["ETH"] = binanceCurrencyInfo{withdrawalPrecision:8}
	c.currenciesInfo["GRS"] = binanceCurrencyInfo{withdrawalPrecision:8}
	c.currenciesInfo["LTC"] = binanceCurrencyInfo{withdrawalPrecision:8}
	c.currenciesInfo["XZC"] = binanceCurrencyInfo{withdrawalPrecision:8}
	return c
}

func (b *Binance) GetName() (string, error) {
	return b.Name, nil
}

func (b *Binance) GetAddress(coin string) (string, error) {
	l.Println("binance - GetAddress - Params() - address :: ", strings.ToLower(coin))
	address, err := b.binanceApi.DepositAddress(binance.AddressRequest{
		Asset:     strings.ToLower(coin),
		Timestamp: time.Now(),
	})

	if err != nil {
		l.Println("binance - GetAddress - DepositAddress() - ", err.Error())
		return "", err
	}

	if address.Address == "" {
		l.Println("binance::GetAddress::EmptyAddress")
		return "", errors.New("empty address")
	}

	return address.Address, nil
}

func (b *Binance) GetDepositStatus(_ string, txId string, _ string) (orderStatus hestia.ExchangeOrderInfo, err error) {
	orderStatus.ReceivedAmount = 0.0
	orderStatus.Status = hestia.ExchangeOrderStatusOpen
	deposits, err := b.binanceApi.DepositHistory(binance.HistoryRequest{
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})
	if err != nil {
		l.Println("binance - GetDepositStatus - DepositHistory() - ", err.Error())
		orderStatus.Status = hestia.ExchangeOrderStatusError
		return
	}
	for _, deposit := range deposits {
		if deposit.TxID == txId {
			switch deposit.Status {
			case 0:
				return
			case 1:
				orderStatus.Status = hestia.ExchangeOrderStatusCompleted
				orderStatus.ReceivedAmount = deposit.Amount
				return
			case 6: // credited but cannot withdraw
				return
			}
		}
	}
	return
}

func (b *Binance) GetWithdrawalTxHash(txId string, asset string) (string, error) {
	withdrawals, err := b.binanceApi.WithdrawHistory(binance.HistoryRequest{
		Asset:     strings.ToLower(asset),
		Timestamp: time.Now(),
	})
	if err != nil {
		l.Println("binance - GetWithdrawalTxHash - WithdrawHistory() - ", err.Error())
		return "", err
	}

	for _, withdrawal := range withdrawals {
		if withdrawal.Id == txId {
			return withdrawal.TxID, nil
		}
	}

	return "", errors.New("binance - GetWithdrawalTxHash() - withdrawal not found")
}

func (b *Binance) GetPair(fromCoin string, toCoin string) (models.TradeInfo, error) {
	var orderSide models.TradeInfo
	fromCoin = strings.ToUpper(fromCoin)
	toCoin = strings.ToUpper(toCoin)

	books, err := b.binanceApi.TickerAllBooks()
	if err != nil {
		l.Println("binance - GetPair - TickerAllBooks() - ", err.Error())
		return orderSide, err
	}
	var bookName string
	for _, book := range books {
		if strings.Contains(book.Symbol, fromCoin) && strings.Contains(book.Symbol, toCoin) {
			bookName = book.Symbol
			break
		}
	}

	fromIndex := strings.Index(bookName, fromCoin)
	toIndex := strings.Index(bookName, toCoin)

	orderSide.Book = bookName
	if fromIndex < toIndex {
		orderSide.Type = "sell"
	} else {
		orderSide.Type = "buy"
	}

	return orderSide, nil
}

func (b *Binance) GetBalance(coin string) (float64, error) {
	res, err := b.binanceApi.Account(binance.AccountRequest{
		RecvWindow: 6 * time.Second,
		Timestamp:  time.Now(),
	})

	if err != nil {
		return 0.0, errors.New("binance - GetBalance - Account() - " + err.Error())
	}
	for _, asset := range res.Balances {
		if asset.Asset == coin {
			return asset.Free, nil
		}
	}
	return 0.0, errors.New("binance - GetBalance - Balances - asset not found " + coin)
}

func (b *Binance) SellAtMarketPrice(order hestia.Trade) (string, error) {
	var side binance.OrderSide
	var newOrder *binance.ProcessedOrder
	var precision int
	var err error
	quoteOrderAvailable, err := b.isQuoteOrderAvailable(order.Symbol)
	if err != nil {
		l.Println("binance - SellAtMarketPrice - isQuoteOrderAvailable() - ", err.Error())
		return "", err
	}
	if order.Side == "buy" && quoteOrderAvailable {
		precision, err = b.getTradePrecision(order.Symbol, "quote")
		if err != nil {
			l.Println("binance - SellAtMarketPrice - getTradePrecision() - " + err.Error())
			return "", err
		}
		order.Amount = roundFixedPrecision(order.Amount, precision)
		side = binance.SideBuy
		newOrder, err = b.binanceApi.NewOrder(binance.NewOrderRequest{
			Symbol:           order.Symbol,
			Side:             side,
			Type:             binance.TypeMarket,
			QuoteOrderQty:    order.Amount,
			Timestamp:        time.Now(),
			NewOrderRespType: binance.RespTypeFull,
		})
	} else {
		var amount float64
		if order.Side == "buy" { // We know that quoteOrders are not available
			avgPrice, err := b.binanceApi.AveragePrice(order.Symbol)
			if err != nil {
				l.Println("binance - SellAtMarketPrice - AveragePrice() - " + err.Error())
				return "", err
			}
			price, err := strconv.ParseFloat(avgPrice.Price, 64)
			if err != nil {
				l.Println("binance - SellAtMarketPrice - ParseFloat()" + err.Error())
				return "", err
			}
			side = binance.SideBuy
			amount = order.Amount / price
		} else {
			side = binance.SideSell
			amount = order.Amount
		}
		precision, err = b.getTradePrecision(order.Symbol, "base")
		if err != nil {
			l.Println("binance - SellAtMarketPrice - getTradePrecision() - " + err.Error())
			return "", err
		}
		amount = roundFixedPrecision(amount, precision)
		newOrder, err = b.binanceApi.NewOrder(binance.NewOrderRequest{
			Symbol:           order.Symbol,
			Side:             side,
			Type:             binance.TypeMarket,
			Quantity:         amount,
			Timestamp:        time.Now(),
			NewOrderRespType: binance.RespTypeFull,
		})
	}
	if err != nil {
		l.Println("binance - SellAtMarketPrice - NewOrder() - " + err.Error())
		return "", err
	}

	return strconv.FormatInt(newOrder.OrderID, 10), nil
}

func (b *Binance) Withdraw(coin string, address string, amount float64) (string, error) {
	amount = roundFixedPrecision(amount, b.currenciesInfo[coin].withdrawalPrecision)
	withdrawal, err := b.binanceApi.Withdraw(binance.WithdrawRequest{
		Asset:      strings.ToLower(coin),
		Address:    address,
		Amount:     amount,
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})

	if err != nil || !withdrawal.Success {
		l.Println(fmt.Sprintf("[Withdraw] Binance failed to withdraw %s", err))
		l.Println(fmt.Sprintf("Msg response: " + withdrawal.Msg))
		if err == nil {
			return "", errors.New(withdrawal.Msg)
		}
		return "", err
	}

	return withdrawal.Id, nil
}

func (b *Binance) GetOrderStatus(order hestia.Trade) (hestia.ExchangeOrderInfo, error) {
	orderId, err := strconv.ParseInt(order.OrderId, 10, 64)
	if err != nil {
		return hestia.ExchangeOrderInfo {
			Status:          hestia.ExchangeOrderStatusError,
			ReceivedAmount:  0,
		}, errors.New("could not parse order id")
	}

	res, err := b.binanceApi.QueryOrder(binance.QueryOrderRequest{
		Symbol:     order.Symbol,
		OrderID:    orderId,
		RecvWindow: 10 * time.Second,
		Timestamp:  time.Now(),
	})

	if err != nil {
		l.Println("binance - GetOrderStatus - QueryOrder - ", err.Error())
		return hestia.ExchangeOrderInfo {
			Status:          hestia.ExchangeOrderStatusError,
			ReceivedAmount:  0,
		}, err
	}
	switch res.Status {
	case binance.StatusFilled:
		return hestia.ExchangeOrderInfo {
			Status:          hestia.ExchangeOrderStatusCompleted,
			ReceivedAmount:  b.getReceivedAmount(*res),
		}, nil
	case binance.StatusNew:
	case binance.StatusPartiallyFilled:
		return hestia.ExchangeOrderInfo{
			Status:          hestia.ExchangeOrderStatusOpen,
		}, nil
	default:
		return hestia.ExchangeOrderInfo{
			Status:          hestia.ExchangeOrderStatusOpen,
		}, errors.New(fmt.Sprintf("binance - GetOrderStatus - unknown/unhandled order status: %s", res.Status))
	}
	return hestia.ExchangeOrderInfo{}, errors.New("status not working")
}

func (b *Binance) isQuoteOrderAvailable(symbol string) (bool, error) {
	info, err := b.binanceApi.ExchangeInfo()
	if err != nil {
		return false, err
	}
	symbol = strings.ToLower(symbol)

	for _, market := range info.Symbols {
		if strings.ToLower(market.Symbol) == symbol {
			return market.QuoteOrderQtyMarketAllowed, nil
		}
	}

	return false, errors.New("symbol not found")
}

func (b *Binance) getTradePrecision(symbol string, option string) (int, error) {
	info, err := b.binanceApi.ExchangeInfo()
	if err != nil {
		return 0, err
	}
	symbol = strings.ToLower(symbol)

	for _, market := range info.Symbols {
		if strings.ToLower(market.Symbol) == symbol {
			if option == "quote" {
				return market.QuotePrecision, nil
			}
			for _, filter := range market.Filters {
				if filter.FilterType == "LOT_SIZE" {
					step, err := strconv.ParseFloat(filter.StepSize, 64)
					if err != nil {
						return 0, err
					}
					return int(-math.Log10(step)), nil
				}
			}
		}
	}

	return 0, errors.New("symbol not found")
}

func (b *Binance) getReceivedAmount(order binance.ExecutedOrder) float64 {
	trades, err := b.binanceApi.MyTrades(binance.MyTradesRequest{
		Symbol:     order.Symbol,
		Limit:      20,
		FromID:     0,
		RecvWindow: 0,
		Timestamp:  time.Now(),
	})
	if err != nil {
		l.Println("binance - Unable to retrieve fees")
		return 0
	}

	fee := 0.0
	for _, trade := range trades {
		if trade.OrderId == int64(order.OrderID) {
			fee = trade.Commission
		}
	}
	var amount float64
	if order.Side == binance.SideBuy {
		amount = order.ExecutedQty - fee
	} else {
		amount = order.CummulativeQuoteQty - fee
	}

	return amount * (1.0 - 0.001) // binance fee
}
