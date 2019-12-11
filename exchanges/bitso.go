package exchanges

import (
	"errors"
	"fmt"
	"github.com/grupokindynos/adrestia-go/exchanges/config"
	"github.com/grupokindynos/adrestia-go/models/balance"
	"github.com/grupokindynos/adrestia-go/models/transaction"
	"github.com/grupokindynos/common/coin-factory/coins"
	"github.com/grupokindynos/common/hestia"
	"github.com/grupokindynos/common/obol"
	bitso "github.com/grupokindynos/gobitso"
	"github.com/grupokindynos/gobitso/models"
	"os"
	"strconv"
	"strings"
)

var BitsoInstance = NewBitso()

type BitsoI struct {
	Exchange
	bitsoService bitso.Bitso
}

func NewBitso() *BitsoI {
	b := new(BitsoI)
	data := b.GetSettings()
	b.bitsoService = *bitso.NewBitso(data.Url)
	b.bitsoService.SetAuth(data.ApiKey, data.ApiSecret)
	return b
}

func (b BitsoI) GetName() (string, error){
	return "bitso", nil
}

func (b BitsoI) GetAddress(coin coins.Coin) (string, error) {
	address, err := b.bitsoService.FundingDestination(models.DestinationParams{FundCurrency:coin.Tag})
	if err != nil {
		return "", err
	}
	return address.Payload.AccountIdentifier, nil
}

func (b BitsoI) SellAtMarketPrice(SellOrder transaction.ExchangeSell) (bool, string, error){
	// TODO Elaborate tests
	bookName, side, err := b.getPair(SellOrder)
	if err != nil {
		return false, "", err
	}
	orderId, err := b.bitsoService.PlaceOrder(models.PlaceOrderParams{
		Book:       bookName,
		Side:       side,
		Type:		"market",
	})
	if err != nil || !orderId.Success {
		return false, "", errors.New("Bitso:SellAtMarketPrice error on request")
	}
	return true, orderId.Payload.Oid, nil
}


func (b *BitsoI) getPair(Order transaction.ExchangeSell) (string, models.OrderSide, error){
	fromCoin := strings.ToLower(Order.FromCoin.Tag)
	toCoin := strings.ToLower(Order.ToCoin.Tag)
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

	if fromIndex < toIndex{
		return bookName, models.Sell, nil
	}
	if toIndex > fromIndex {
		return bookName, models.Buy, nil
	}
	return bookName, "unknown", errors.New("could not find a satisfying book")
}

func (b BitsoI) OneCoinToBtc(coin coins.Coin) (float64, error) {
	if coin.Tag == "BTC" {
		return 1.0, nil
	}
	rate, err := obol.GetCoin2CoinRatesWithAmount("https://obol-rates.herokuapp.com/", "btc", coin.Tag, fmt.Sprintf("%f", 1.0))
	if err != nil {
		return 0.0, err
	}
	return rate.AveragePrice, nil
}

func (b BitsoI) GetOrderStatus(orderId string) (status hestia.AdrestiaStatus, err error){
	var wrappedOrder []string
	wrappedOrder = append(wrappedOrder, orderId)
	res, err := b.bitsoService.LookUpOrders(wrappedOrder)
	if err != nil {
		return
	}
	if res.Payload[0].Status == "completed" {
		return hestia.AdrestiaStatusCompleted, nil
	}
	if res.Payload[0].Status == "partial-fill" {
		return hestia.AdrestiaStatusPartiallyFulfilled, nil
	}
	if res.Payload[0].Status == "open" || res.Payload[0].Status == "queued" {
		return hestia.AdrestiaStatusCreated, nil
	}
	return hestia.AdrestiaStatusCreated, errors.New("unknown order status " + res.Payload[0].Status)
}

func (b BitsoI) GetBalances() ([]balance.Balance, error) {
	bal, err := b.bitsoService.Balances()
	if err!=nil {
		return nil, err
	}
	var balances []balance.Balance

	for _, asset := range bal.Payload.Balances {
		rate, _ := obol.GetCoin2CoinRates("https://obol-rates.herokuapp.com/", "BTC", asset.Currency)
		confirmedAmount, err := strconv.ParseFloat(asset.Available, 64)
		unconfirmedAmount, err := strconv.ParseFloat(asset.Available, 64)
		if err != nil {
			return nil, err
		}
		var b = balance.Balance{
			Ticker:     asset.Currency,
			ConfirmedBalance:    	confirmedAmount,
			UnconfirmedBalance: 	unconfirmedAmount,
			RateBTC:    			rate,
			DiffBTC:    			0,
			IsBalanced: 			false,
		}
		if b.GetTotalBalance() > 0.0 {
			balances = append(balances, b)
		}
	}
	fmt.Println("Balances BitsoI: ", balances)
	return balances, nil
}

func (b BitsoI) GetSettings() config.BitsoAuth {
	var data config.BitsoAuth
	data.ApiKey = os.Getenv("BITSO_API_KEY")
	data.ApiSecret = os.Getenv("BITSO_API_SECRET")
	data.Url = os.Getenv("BITSO_URL")
	return data
}
