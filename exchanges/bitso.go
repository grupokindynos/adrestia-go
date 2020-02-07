package exchanges

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

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

func (b *Bitso) SellAtMarketPrice(sellOrder hestia.ExchangeOrder) (string, error) {
	var side models.OrderSide
	orderSide, err := b.GetPair(sellOrder.SoldCurrency, sellOrder.ReceivedCurrency)
	if err != nil {
		return "", err
	}

	if orderSide.Type == "buy" {
		side = models.Buy
	} else {
		side = models.Sell
	}

	orderId, err := b.bitsoService.PlaceOrder(models.PlaceOrderParams{
		Book: orderSide.Book,
		Side: side,
		Type: "market",
	})
	if err != nil || !orderId.Success {
		log.Println("Error::Bitso::SellAtMarketPrice::", err)
		return "", errors.New("Bitso:SellAtMarketPrice error on request: ")
	}
	return orderId.Payload.Oid, nil
}

func (b *Bitso) Withdraw(coin coins.Coin, address string, amount float64) (string, error) {
	res, err := b.bitsoService.CryptoWithdrawal(models.WithdrawParams{
		Currency: coin.Info.Tag,
		Amount:   fmt.Sprintf("%f", amount),
		Address:  address,
		Tag:      "adrestia balancing - " + time.Now().String(),
	})

	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", res.Payload.Details.TxHash), nil
}

func (b *Bitso) GetRateByAmount(sell transaction.ExchangeSell) (float64, error) {
	return 0.0, errors.New("func not implemented")
}

func (b *Bitso) GetWithdrawalTxHash(txId string, asset string, address string, withdrawalAmount float64) (string, error) {
	return txId, nil
}

func (b *Bitso) GetDepositStatus(txId string, asset string) (hestia.OrderStatus, error) {
	var status hestia.OrderStatus
	deposits, err := b.bitsoService.Fundings(models.FundingParams{})
	if err != nil {
		log.Println("135")
		return hestia.OrderStatus{}, err
	}
	if deposits.Success != true {
		return hestia.OrderStatus{}, errors.New("Response not succesful")
	}

	status.Status = hestia.ExchangeStatusError
	for _, deposit := range deposits.Payload {
		if deposit.Details.TxHash == txId {
			if deposit.Status == "completed" {
				status.Status = hestia.ExchangeStatusCompleted
				amount, err := strconv.ParseFloat(deposit.Amount, 64)
				if err != nil {
					return hestia.OrderStatus{}, err
				}
				status.AvailableAmount = amount
				return status, nil
			} else if deposit.Status == "partial-fill" || deposit.Status == "open" || deposit.Status == "queued" {
				status.Status = hestia.ExchangeStatusOpen
				return status, nil
			} else {
				return status, errors.New("unkown deposit status " + deposit.Status)
			}
		}
	}

	return status, errors.New("deposit not found")
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

func (b *Bitso) GetPair(fromCoin string, toCoin string) (OrderSide, error) {
	var orderSide OrderSide
	fromCoin = strings.ToLower(fromCoin)
	toCoin = strings.ToLower(toCoin)
	books, err := b.bitsoService.AvailableBooks()
	if err != nil {
		return orderSide, err
	}
	var bookName string
	for _, book := range books.Payload {
		if strings.Contains(book.Book, fromCoin) && strings.Contains(book.Book, toCoin) {
			bookName = book.Book
			break
		}
	}
	// ltc_btc
	fromIndex := strings.Index(bookName, fromCoin)
	toIndex := strings.Index(bookName, toCoin)

	orderSide.Book = bookName
	// check bitso convention
	if fromIndex < toIndex {
		orderSide.Type = "sell"
		orderSide.ReceivedCurrency = toCoin
		orderSide.SoldCurrency = fromCoin
	} else {
		orderSide.Type = "buy"
		orderSide.ReceivedCurrency = fromCoin
		orderSide.SoldCurrency = toCoin
	}

	return orderSide, nil
}
