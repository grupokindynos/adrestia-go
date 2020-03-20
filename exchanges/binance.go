package exchanges

import (
	"context"
	"errors"
	"fmt"
	l "log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	"github.com/grupokindynos/go-binance"
	"github.com/joho/godotenv"
)

type Binance struct {
	Exchange
	AccountName string
	binanceApi  binance.Binance
	Obol        obol.ObolService
}

func NewBinance(params Params) *Binance {
	c := new(Binance)
	c.Name = "binance"
	c.BaseUrl = ""
	data := GetSettings()
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	hmacSigner := &binance.HmacSigner{
		Key: []byte(data.PrivateApi),
	}
	ctx, _ := context.WithCancel(context.Background())
	binanceService := binance.NewAPIService(
		"https://api.binance.com",
		data.PublicApi,
		hmacSigner,
		logger,
		ctx,
	)
	c.binanceApi = binance.NewBinance(binanceService)
	return c
}

func (b *Binance) GetName() (string, error) {
	return b.Name, nil
}

func (b *Binance) GetAddress(coin coins.Coin) (string, error) {
	address, err := b.binanceApi.DepositAddress(binance.AddressRequest{
		Asset:     strings.ToLower(coin.Info.Tag),
		Timestamp: time.Now(),
	})
	if err != nil {
		l.Println("binance - GetAddress - DepositAddress() - ", err.Error())
		return "", err
	}
	return address.Address, nil
}

func (b *Binance) GetDepositStatus(txid string, asset string) (orderStatus hestia.OrderStatus, err error) {
	orderStatus.AvailableAmount = 0.0
	orderStatus.Status = hestia.ExchangeStatusOpen
	deposits, err := b.binanceApi.DepositHistory(binance.HistoryRequest{
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})
	if err != nil {
		l.Println("binance - GetDepositStatus - DepositHistory() - ", err.Error())
		orderStatus.Status = hestia.ExchangeStatusError
		return
	}
	for _, deposit := range deposits {
		if deposit.TxID == txid {
			switch deposit.Status {
			case 0:
				return
			case 1:
				orderStatus.Status = hestia.ExchangeStatusCompleted
				orderStatus.AvailableAmount = deposit.Amount
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

func (b *Binance) GetPair(fromCoin string, toCoin string) (OrderSide, error) {
	var orderSide OrderSide
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
		orderSide.ReceivedCurrency = toCoin
		orderSide.SoldCurrency = fromCoin
	} else {
		orderSide.Type = "buy"
		orderSide.ReceivedCurrency = toCoin
		orderSide.SoldCurrency = fromCoin
	}

	return orderSide, nil
}

func (b *Binance) GetBalance(coin string) (float64, error) {
	res, err := b.binanceApi.Account(binance.AccountRequest{
		RecvWindow: 5 * time.Second,
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
	return 0.0, errors.New("binance - GetBalance - Balances - " + err.Error())
}

func (b *Binance) SellAtMarketPrice(order hestia.ExchangeOrder) (string, error) {
	var side binance.OrderSide
	var newOrder *binance.ProcessedOrder
	var err error
	quoteOrderAvailable, err := b.isQuoteOrderAvailable(order.Symbol)
	if err != nil {
		l.Println("binance - SellAtMarketPrice - isQuoteOrderAvailable() - ", err.Error())
		return "", err
	}
	if order.Side == "buy" && quoteOrderAvailable {
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
		precision, err := b.getTradePrecision(order.Symbol)
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

func (b *Binance) Withdraw(coin coins.Coin, address string, amount float64) (string, error) {
	withdrawal, err := b.binanceApi.Withdraw(binance.WithdrawRequest{
		Asset:      strings.ToLower(coin.Info.Tag),
		Address:    address,
		Amount:     amount,
		Name:       "Adrestia-go Withdrawal",
		RecvWindow: 6 * time.Second,
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

func (b *Binance) GetOrderStatus(order hestia.ExchangeOrder) (hestia.OrderStatus, error) {
	orderId, err := strconv.ParseInt(order.OrderId, 10, 64)
	if err != nil {
		return hestia.OrderStatus{
			Status:          hestia.ExchangeStatusError,
			AvailableAmount: 0,
		}, errors.New("could not parse order id")
	}

	res, err := b.binanceApi.QueryOrder(binance.QueryOrderRequest{
		Symbol:     order.Symbol,
		OrderID:    orderId,
		RecvWindow: 5 * time.Second,
		Timestamp:  time.Now(),
	})

	if err != nil {
		l.Println("binance - GetOrderStatus - QueryOrder - ", err.Error())
		return hestia.OrderStatus{
			Status:          hestia.ExchangeStatusError,
			AvailableAmount: 0,
		}, err
	}

	switch res.Status {
	case binance.StatusFilled:
		return hestia.OrderStatus{
			Status:          hestia.ExchangeStatusCompleted,
			AvailableAmount: b.getReceivedAmount(*res),
		}, nil
	case binance.StatusNew:
		return hestia.OrderStatus{
			Status:          hestia.ExchangeStatusOpen,
			AvailableAmount: 0,
		}, nil
	case binance.StatusPartiallyFilled:
		return hestia.OrderStatus{
			Status:          hestia.ExchangeStatusOpen,
			AvailableAmount: 0,
		}, nil
	default:
		return hestia.OrderStatus{
			Status:          hestia.ExchangeStatusOpen,
			AvailableAmount: 0,
		}, errors.New(fmt.Sprintf("binance - GetOrderStatus - unknown/unhandled order status: %s", res.Status))
	}
}

func GetSettings() config.BinanceAuth {
	if err := godotenv.Load(); err != nil {
		l.Println(err)
	}
	var data config.BinanceAuth
	data.PublicApi = os.Getenv("BINANCE_PUB_API")
	data.PrivateApi = os.Getenv("BINANCE_PRIV_API")
	return data
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

func (b *Binance) getTradePrecision(symbol string) (int, error) {
	info, err := b.binanceApi.ExchangeInfo()
	if err != nil {
		return 0, err
	}
	symbol = strings.ToLower(symbol)

	for _, market := range info.Symbols {
		if strings.ToLower(market.Symbol) == symbol {
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

func round(f float64) float64 {
	return math.Floor(f + .5)
}

func roundFixedPrecision(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return round(f*shift) / shift
}

func (b *Binance) getReceivedAmount(order binance.ExecutedOrder) float64 {
	if order.Side == binance.SideBuy {
		return order.ExecutedQty
	} else {
		return order.CummulativeQuoteQty
	}
}