package exchanges

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/grupokindynos/adrestia-go/exchanges/config"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	bitso "github.com/grupokindynos/gobitso"
	"github.com/grupokindynos/gobitso/models"
)

type Bitso struct {
	Exchange
	bitsoService bitso.Bitso
	Obol         obol.ObolService
}

func NewBitso(params Params) *Bitso {
	b := new(Bitso)
	data := b.getSettings()
	b.bitsoService = *bitso.NewBitso(data.Url)
	b.bitsoService.SetAuth(data.ApiKey, data.ApiSecret)
	b.Obol = params.Obol
	return b
}

func (b *Bitso) GetName() (string, error) {
	return "bitso", nil
}

func (b *Bitso) GetAddress(coin coins.Coin) (string, error) {
	address, err := b.bitsoService.FundingDestination(models.DestinationParams{FundCurrency: coin.Info.Tag})
	if err != nil {
		return "", err
	}
	return address.Payload.AccountIdentifier, nil
}

func (b *Bitso) OneCoinToBtc(coin coins.Coin) (float64, error) {
	if coin.Info.Tag == "BTC" {
		return 1.0, nil
	}
	rate, err := b.Obol.GetCoin2CoinRatesWithAmount("btc", coin.Info.Tag, fmt.Sprintf("%f", 1.0))
	if err != nil {
		return 0.0, err
	}
	return rate.AveragePrice, nil
}

func (b *Bitso) GetBalances() ([]balance.Balance, error) {
	bal, err := b.bitsoService.Balances()
	if err != nil {
		return nil, err
	}
	var balances []balance.Balance

	for _, asset := range bal.Payload.Balances {
		rate, _ := b.Obol.GetCoin2CoinRates("BTC", asset.Currency)
		confirmedAmount, err := strconv.ParseFloat(asset.Available, 64)
		unconfirmedAmount, err := strconv.ParseFloat(asset.Available, 64)
		if err != nil {
			return nil, err
		}
		var b = balance.Balance{
			Ticker:             asset.Currency,
			ConfirmedBalance:   confirmedAmount,
			UnconfirmedBalance: unconfirmedAmount,
			RateBTC:            rate,
			DiffBTC:            0,
			IsBalanced:         false,
		}
		if b.GetTotalBalance() > 0.0 {
			balances = append(balances, b)
		}
	}
	fmt.Println("Balances BitsoI: ", balances)
	return balances, nil
}

func (b *Bitso) SellAtMarketPrice(sellOrder transaction.ExchangeSell) (bool, string, error) {
	// TODO Elaborate tests
	bookName, side, err := b.getPair(sellOrder)
	if err != nil {
		return false, "", err
	}
	orderId, err := b.bitsoService.PlaceOrder(models.PlaceOrderParams{
		Book: bookName,
		Side: side,
		Type: "market",
	})
	if err != nil || !orderId.Success {
		return false, "", errors.New("Bitso:SellAtMarketPrice error on request")
	}
	return true, orderId.Payload.Oid, nil
}

func (b *Bitso) Withdraw(coin coins.Coin, address string, amount float64) (bool, error) {
	return false, errors.New("func not implemented")
}

func (b *Bitso) GetRateByAmount(sell transaction.ExchangeSell) (float64, error) {
	return 0.0, errors.New("func not implemented")
}

func (b *Bitso) GetOrderStatus(order hestia.ExchangeOrder) (status hestia.OrderStatus, err error) {
	var wrappedOrder []string
	wrappedOrder = append(wrappedOrder, order.OrderId)
	res, err := b.bitsoService.LookUpOrders(wrappedOrder)
	if err != nil {
		return
	}
	if res.Payload[0].Status == "completed" {
		amount, _ := strconv.ParseFloat(res.Payload[0].OriginalValue, 64) // TODO Verify this is correct in API
		return hestia.OrderStatus{
			Status:          hestia.ExchangeStatusCompleted,
			AvailableAmount: amount,
		}, nil
	}
	if res.Payload[0].Status == "partial-fill" {
		return hestia.OrderStatus{
			Status:          hestia.ExchangeStatusOpen,
			AvailableAmount: 0,
		}, nil
	}
	if res.Payload[0].Status == "open" || res.Payload[0].Status == "queued" {
		return hestia.OrderStatus{
			Status:          hestia.ExchangeStatusOpen,
			AvailableAmount: 0,
		}, nil
	}
	return hestia.OrderStatus{
		Status:          hestia.ExchangeStatusError,
		AvailableAmount: 0,
	}, errors.New("unknown order status " + res.Payload[0].Status)
}

func (b *Bitso) getSettings() config.BitsoAuth {
	var data config.BitsoAuth
	data.ApiKey = os.Getenv("BITSO_API_KEY")
	data.ApiSecret = os.Getenv("BITSO_API_SECRET")
	data.Url = os.Getenv("BITSO_URL")
	return data
}

func (b *Bitso) getPair(Order transaction.ExchangeSell) (string, models.OrderSide, error) {
	fromCoin := strings.ToLower(Order.FromCoin.Info.Tag)
	toCoin := strings.ToLower(Order.ToCoin.Info.Tag)
	books, err := b.bitsoService.AvailableBooks()
	if err != nil {
		return "", "", err
	}
	var bookName string
	for _, book := range books.Payload {
		if strings.Contains(book.Book, fromCoin) && strings.Contains(book.Book, toCoin) {
			bookName = book.Book
		}
	}
	// ltc_btc
	fromIndex := strings.Index(bookName, fromCoin)
	toIndex := strings.Index(bookName, toCoin)

	if fromIndex < toIndex {
		return bookName, models.Sell, nil
	}
	if toIndex > fromIndex {
		return bookName, models.Buy, nil
	}
	return bookName, "unknown", errors.New("could not find a satisfying book")
}
